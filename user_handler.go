package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/labstack/echo/v4"
	"github.com/mojura/enkodo"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultSessionIDKey      = "SESSIONID"
	defaultSessionExpiresKey = "EXPIRES"
	defaultUserIDKey         = "USERID"
	defaultUsernameKey       = "USERNAME"
	bcryptDefaultCost        = bcrypt.MinCost
)

var (
	fallbackImage = "../img/NoImage.jpg"
	userLock      sync.RWMutex
	userCache     map[int64]UserModel
	userNameCache map[string]int64
)

type UserModel struct {
	ID             int64  `db:"id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	Description    string `db:"description"`
	HashedPassword string `db:"password"`
	IconHash       string `db:"icon_hash"`
	DarkMode       bool   `db:"dark_mode"`
}

type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Theme       Theme  `json:"theme,omitempty"`
	IconHash    string `json:"icon_hash,omitempty"`
}

type Theme struct {
	ID       int64 `json:"id"`
	DarkMode bool  `json:"dark_mode"`
}

type ThemeModel struct {
	ID       int64 `db:"id"`
	UserID   int64 `db:"user_id"`
	DarkMode bool  `db:"dark_mode"`
}

type PostUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	// Password is non-hashed password.
	Password string               `json:"password"`
	Theme    PostUserRequestTheme `json:"theme"`
}

type PostUserRequestTheme struct {
	DarkMode bool `json:"dark_mode"`
}

type LoginRequest struct {
	Username string `json:"username"`
	// Password is non-hashed password.
	Password string `json:"password"`
}

type PostIconRequest struct {
	Image []byte `json:"image"`
}

type PostIconResponse struct {
	ID int64 `json:"id"`
}

func (m UserModel) toUser() User {
	if m.IconHash == "" {
		m.IconHash = "d9f8294e9d895f81ce62e73dc7d5dff862a4fa40bd4e0fecf53f7526a8edcac0"
	}
	user := User{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Description: m.Description,
		Theme: Theme{
			ID:       m.ID,
			DarkMode: m.DarkMode,
		},
		IconHash: m.IconHash,
	}
	return user
}

func getIconHandler(c echo.Context) error {

	username := c.Param("username")

	user, exists := getUserByName(username)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "user not found: "+username)
	}

	match, ok := c.Request().Header["If-None-Match"]
	if ok && strings.Contains(match[0], user.IconHash) {
		return c.NoContent(http.StatusNotModified)
	}

	var image []byte
	if err := dbConn.GetContext(context.Background(), &image, "SELECT image FROM icons WHERE user_id = ?", user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.File(fallbackImage)
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user icon: "+err.Error())
		}
	}

	return c.Blob(http.StatusOK, "image/jpeg", image)
}

func UnsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func getIconFiber(c *fiber.Ctx) error {

	username := c.Params("username")
	user, exists := getUserByName(username)
	if !exists {
		return c.Status(fasthttp.StatusNotFound).SendString("user not found: " + username)
	}

	noneMatch := c.Request().Header.Peek(fasthttp.HeaderIfNoneMatch)
	if len(noneMatch) > 0 && bytes.Contains(noneMatch, UnsafeBytes(user.IconHash)) {
		return c.SendStatus(fasthttp.StatusNotModified)
	}

	var image []byte
	if err := dbConn.GetContext(c.Context(), &image, "SELECT image FROM icons WHERE user_id = ?", user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.SendFile(fallbackImage)
		} else {
			return c.Status(fasthttp.StatusInternalServerError).SendString("failed to get user icon: " + err.Error())
		}
	}

	return c.Status(fasthttp.StatusOK).Type("image/jpeg").Send(image)
}

func postIconHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	var req *PostIconRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}
	iconHash := fmt.Sprintf("%x", sha256.Sum256(req.Image))

	if _, err := dbConn.ExecContext(ctx, "UPDATE users SET icon_hash = ? WHERE id = ?", iconHash, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete old user icon: "+err.Error())
	}

	if _, err := dbConn.ExecContext(ctx, "DELETE FROM icons WHERE user_id = ?", userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete old user icon: "+err.Error())
	}

	rs, err := dbConn.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", userID, req.Image)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert new user icon: "+err.Error())
	}

	iconID, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get last inserted icon id: "+err.Error())
	}

	userLock.Lock()
	if u, ok := userCache[userID]; ok {
		u.IconHash = iconHash
		userCache[userID] = u
	}
	userLock.Unlock()

	return c.JSON(http.StatusCreated, &PostIconResponse{
		ID: iconID,
	})
}

func getMeHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	userModel := UserModel{}
	err := dbConn.GetContext(ctx, &userModel, "SELECT * FROM users WHERE id = ?", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the userid in session")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	user, err := fillUserResponse(ctx, dbConn, userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill user: "+err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// ユーザ登録API
// POST /api/register
func registerHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := PostUserRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	if req.Name == "pipe" {
		return echo.NewHTTPError(http.StatusBadRequest, "the username 'pipe' is reserved")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptDefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate hashed password: "+err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil) // post
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	userModel := UserModel{
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		HashedPassword: string(hashedPassword),
		DarkMode:       req.Theme.DarkMode,
	}

	result, err := tx.NamedExecContext(ctx, "INSERT INTO users (name, display_name, description, password, dark_mode) VALUES(:name, :display_name, :description, :password, :dark_mode)", userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert user: "+err.Error())
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get last inserted user id: "+err.Error())
	}

	userModel.ID = userID

	user, err := fillUserResponse(ctx, tx, userModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill user: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	if out, err := exec.Command("pdnsutil", "add-record", "u.isucon.dev", req.Name, "A", "30", powerDNSSubdomainAddress).CombinedOutput(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, string(out)+": "+err.Error())
	}

	userLock.Lock()
	userCache[userModel.ID] = userModel
	userNameCache[userModel.Name] = userModel.ID
	userLock.Unlock()

	return c.JSON(http.StatusCreated, user)
}

// ユーザログインAPI
// POST /api/login
func loginHandler(c echo.Context) error {
	//ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := LoginRequest{}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	userModel, exists := getUserByName(req.Username)
	if !exists {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(userModel.HashedPassword), []byte(req.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to compare hash and password: "+err.Error())
	}

	sess := getSession(c)
	sess.Values.UserID = userModel.ID
	sess.Values.UserName = userModel.Name
	sess.Save(c)

	return c.NoContent(http.StatusOK)
}

// ユーザ詳細API
// GET /api/user/:username
func getUserHandler(c echo.Context) error {
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")

	userModel, exists := getUserByName(username)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}

	user := userModel.toUser()

	return c.JSON(http.StatusOK, user)
}

func verifyUserSession(c echo.Context) error {
	sess := getSession(c)
	userID := sess.Values.UserID
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get USERID value from session")
	}
	return nil
}

func fillUserResponse(ctx context.Context, tx dbtx, userModel UserModel) (User, error) {
	return userModel.toUser(), nil
}

func getUserMap(ctx context.Context, tx dbtx, userIDs []int64) (map[int64]UserModel, error) {
	userMap := map[int64]UserModel{}
	if len(userIDs) == 0 {
		return userMap, nil
	}
	slices.Sort(userIDs)
	userIDs = slices.Compact(userIDs)
	userLock.RLock()
	defer userLock.RUnlock()
	for _, i := range userIDs {
		if user, ok := userCache[i]; ok {
			userMap[i] = user
		}
	}
	return userMap, nil
}

func getUserByName(name string) (UserModel, bool) {
	userLock.RLock()
	defer userLock.RUnlock()
	var userID int64
	var user UserModel
	var ok bool
	if userID, ok = userNameCache[name]; !ok {
		return UserModel{}, false
	}
	if user, ok = userCache[userID]; !ok {
		return UserModel{}, false
	}
	return user, true
}

func getUserByID(userID int64) (UserModel, bool) {
	userLock.RLock()
	defer userLock.RUnlock()
	var user UserModel
	var ok bool
	if user, ok = userCache[userID]; !ok {
		return UserModel{}, false
	}
	return user, true
}

func getUserAll() []UserModel {
	userLock.RLock()
	defer userLock.RUnlock()
	users := make([]UserModel, len(userCache))
	for _, u := range userCache {
		users = append(users, u)
	}
	return users
}

func warmupUsersCache(ctx context.Context) {
	uCache := map[int64]UserModel{}
	nCache := map[string]int64{}
	users := []UserModel{}
	query := `SELECT * FROM users`
	if err := dbConn.SelectContext(ctx, &users, query); err != nil {
		return
	}
	for _, u := range users {
		uCache[u.ID] = u
		nCache[u.Name] = u.ID
	}
	userLock.Lock()
	userCache = uCache
	userNameCache = nCache
	userLock.Unlock()
}

type simpleCookie struct {
	Values sessionData
	Key    string
}

type sessionData struct {
	UserID   int64
	UserName string
}

func (s *sessionData) MarshalEnkodo(enc *enkodo.Encoder) error {
	enc.Int64(s.UserID)
	enc.String(s.UserName)
	return nil
}

func (s *sessionData) UnmarshalEnkodo(dec *enkodo.Decoder) error {
	var err error
	if s.UserID, err = dec.Int64(); err != nil {
		return err
	}

	if s.UserName, err = dec.String(); err != nil {
		return err
	}

	return nil
}

func (s *simpleCookie) Save(c echo.Context) error {
	bs, err := enkodo.Marshal(&s.Values)
	if err != nil {
		return err
	}
	cookie := new(http.Cookie)
	cookie.Name = defaultSessionIDKey
	cookie.Value = base64.StdEncoding.EncodeToString(bs)
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	cookie.Domain = "u.isucon.dev"
	c.SetCookie(cookie)
	return nil
}

func getSession(c echo.Context) *simpleCookie {
	session, err := readSession(c, defaultSessionIDKey)
	if err != nil {
		log.Print(err)
	}
	return session
}

func readSession(c echo.Context, key string) (*simpleCookie, error) {
	s := c.Get(key)
	if s != nil {
		return s.(*simpleCookie), nil
	}
	var sd sessionData
	cookies, err := c.Cookie(key)
	if err == nil {
		if cookie := cookies.Value; cookie != "" {
			data, err := base64.StdEncoding.DecodeString(cookie)
			if err != nil {
				return nil, err
			}
			if err := enkodo.Unmarshal(data, &sd); err != nil {
				return nil, err
			}
		}
	}
	newSession := &simpleCookie{
		Values: sd,
		Key:    key,
	}
	c.Set(key, newSession)
	return newSession, nil
}
