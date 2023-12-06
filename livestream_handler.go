package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

type ReserveLivestreamRequest struct {
	Tags         []int64 `json:"tags"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	PlaylistUrl  string  `json:"playlist_url"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	StartAt      int64   `json:"start_at"`
	EndAt        int64   `json:"end_at"`
}

type LivestreamViewerModel struct {
	UserID       int64 `db:"user_id" json:"user_id"`
	LivestreamID int64 `db:"livestream_id" json:"livestream_id"`
	CreatedAt    int64 `db:"created_at" json:"created_at"`
}

type LivestreamModel struct {
	ID           int64  `db:"id" json:"id"`
	UserID       int64  `db:"user_id" json:"user_id"`
	Title        string `db:"title" json:"title"`
	Description  string `db:"description" json:"description"`
	PlaylistUrl  string `db:"playlist_url" json:"playlist_url"`
	ThumbnailUrl string `db:"thumbnail_url" json:"thumbnail_url"`
	StartAt      int64  `db:"start_at" json:"start_at"`
	EndAt        int64  `db:"end_at" json:"end_at"`
	RawTags      string `db:"raw_tags"`
}

type Livestream struct {
	ID           int64  `json:"id"`
	Owner        User   `json:"owner"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	PlaylistUrl  string `json:"playlist_url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	Tags         []Tag  `json:"tags"`
	StartAt      int64  `json:"start_at"`
	EndAt        int64  `json:"end_at"`
}

type LivestreamTagModel struct {
	ID           int64 `db:"id" json:"id"`
	LivestreamID int64 `db:"livestream_id" json:"livestream_id"`
	TagID        int64 `db:"tag_id" json:"tag_id"`
}

type ReservationSlotModel struct {
	ID      int64 `db:"id" json:"id"`
	Slot    int64 `db:"slot" json:"slot"`
	StartAt int64 `db:"start_at" json:"start_at"`
	EndAt   int64 `db:"end_at" json:"end_at"`
}

var TagMap = map[int64]Tag{}
var TagIDMap = map[string]int64{}
var Tags = []string{"ライブ配信", "ゲーム実況", "生放送", "アドバイス", "初心者歓迎", "プロゲーマー", "新作ゲーム", "レトロゲーム", "RPG", "FPS", "アクションゲーム", "対戦ゲーム", "マルチプレイ", "シングルプレイ", "ゲーム解説", "ホラーゲーム", "イベント生放送", "新情報発表", "Q&Aセッション", "チャット交流", "視聴者参加", "音楽ライブ", "カバーソング", "オリジナル楽曲", "アコースティック", "歌配信", "楽器演奏", "ギター", "ピアノ", "バンドセッション", "DJセット", "トーク配信", "朝活", "夜ふかし", "日常話", "趣味の話", "語学学習", "お料理配信", "手料理", "レシピ紹介", "アート配信", "絵描き", "DIY", "手芸", "アニメトーク", "映画レビュー", "読書感想", "ファッション", "メイク", "ビューティー", "健康", "ワークアウト", "ヨガ", "ダンス", "旅行記", "アウトドア", "キャンプ", "ペットと一緒", "猫", "犬", "釣り", "ガーデニング", "テクノロジー", "ガジェット紹介", "プログラミング", "DIY電子工作", "ニュース解説", "歴史", "文化", "社会問題", "心理学", "宇宙", "科学", "マジック", "コメディ", "スポーツ", "サッカー", "野球", "バスケットボール", "ライフハック", "教育", "子育て", "ビジネス", "起業", "投資", "仮想通貨", "株式投資", "不動産", "キャリア", "スピリチュアル", "占い", "手相", "オカルト", "UFO", "都市伝説", "コンサート", "ファンミーティング", "コラボ配信", "記念配信", "生誕祭", "周年記念", "サプライズ", "椅子"}

var (
	livestreamLock  sync.RWMutex
	livestreamCache map[int64]LivestreamModel
)

func init() {
	for i, t := range Tags {
		TagMap[int64(i+1)] = Tag{int64(i + 1), t}
		TagIDMap[t] = int64(i + 1)
	}
}

func reserveLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	var req *ReserveLivestreamRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	tx, err := dbConn.BeginTxx(ctx, nil) // post
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	// 2023/11/25 10:00からの１年間の期間内であるかチェック
	var (
		termStartAt    = time.Date(2023, 11, 25, 1, 0, 0, 0, time.UTC)
		termEndAt      = time.Date(2024, 11, 25, 1, 0, 0, 0, time.UTC)
		reserveStartAt = time.Unix(req.StartAt, 0)
		reserveEndAt   = time.Unix(req.EndAt, 0)
	)
	if (reserveStartAt.Equal(termEndAt) || reserveStartAt.After(termEndAt)) || (reserveEndAt.Equal(termStartAt) || reserveEndAt.Before(termStartAt)) {
		return echo.NewHTTPError(http.StatusBadRequest, "bad reservation time range")
	}

	startIndex := math.Ceil(float64(req.StartAt-1700874000) / 3600)
	endIndex := math.Floor(float64(req.EndAt-1700874000) / 3600)

	// 予約枠をみて、予約が可能か調べる
	// NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
	//var slots []*ReservationSlotModel
	var minslot int64
	if err := tx.GetContext(ctx, &minslot, "SELECT MIN(slot) FROM reservation_slots WHERE id >= ? AND id <= ? FOR UPDATE", startIndex, endIndex); err != nil {
		c.Logger().Warnf("予約枠一覧取得でエラー発生: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get reservation_slots: "+err.Error())
	}
	if minslot < 1 {
		return c.NoContent(http.StatusBadRequest)
		// return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("予約期間 %d ~ %dに対して、予約区間 %d(%f) ~ %d(%f)が予約できません", termStartAt.Unix(), termEndAt.Unix(), req.StartAt, startIndex, req.EndAt, endIndex))
	}
	/*
		for _, slot := range slots {
			var count int
			if err := tx.GetContext(ctx, &count, "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?", slot.StartAt, slot.EndAt); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get reservation_slots: "+err.Error())
			}
			c.Logger().Infof("%d ~ %d予約枠の残数 = %d\n", slot.StartAt, slot.EndAt, slot.Slot)
			if count < 1 {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("予約期間 %d ~ %dに対して、予約区間 %d ~ %dが予約できません", termStartAt.Unix(), termEndAt.Unix(), req.StartAt, req.EndAt))
			}
		}
	*/

	tagIDs := []string{}
	for _, tagID := range req.Tags {
		tagIDs = append(tagIDs, strconv.Itoa(int(tagID)))
	}

	var (
		livestreamModel = &LivestreamModel{
			UserID:       int64(userID),
			Title:        req.Title,
			Description:  req.Description,
			PlaylistUrl:  req.PlaylistUrl,
			ThumbnailUrl: req.ThumbnailUrl,
			StartAt:      req.StartAt,
			EndAt:        req.EndAt,
			RawTags:      strings.Join(tagIDs, ","),
		}
	)

	if _, err := tx.ExecContext(ctx, "UPDATE reservation_slots SET slot = slot - 1 WHERE id >= ? AND id <= ?", startIndex, endIndex); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update reservation_slot: "+err.Error())
	}

	rs, err := tx.NamedExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at, raw_tags) VALUES(:user_id, :title, :description, :playlist_url, :thumbnail_url, :start_at, :end_at, :raw_tags)", livestreamModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert livestream: "+err.Error())
	}

	livestreamID, err := rs.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get last inserted livestream id: "+err.Error())
	}
	livestreamModel.ID = livestreamID

	// タグ追加
	if len(req.Tags) > 0 {
		values := []string{}
		params := []any{}
		for _, tagID := range req.Tags {
			values = append(values, "(?, ?)")
			params = append(params, livestreamID, tagID)
		}
		query := `INSERT INTO livestream_tags (livestream_id, tag_id) VALUES ` + strings.Join(values, ", ")
		_, err = tx.ExecContext(ctx, query, params...)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert tags: "+err.Error())
		}
	}

	userMap, err := getUserMap(ctx, tx, []int64{userID})
	if err != nil {
		return err
	}
	livestream, err := fillLivestreamResponse(ctx, tx, *livestreamModel, userMap)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
	}

	livestreamLock.Lock()
	livestreamCache[livestreamModel.ID] = *livestreamModel
	livestreamLock.Unlock()

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit: "+err.Error())
	}

	return c.JSON(http.StatusCreated, livestream)
}

func searchLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	keyTagName := c.QueryParam("tag")

	var livestreamIDs []int64
	if c.QueryParam("tag") != "" {
		// タグによる取得
		tagID, ok := TagIDMap[keyTagName]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		if err := dbConn.SelectContext(ctx, &livestreamIDs, "SELECT l.id FROM livestreams l JOIN livestream_tags lt ON lt.livestream_id = l.id WHERE tag_id=? GROUP BY l.id ORDER BY l.id DESC", tagID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
		}
	} else {
		// 検索条件なし
		query := `SELECT l.id FROM livestreams l ORDER BY id DESC`
		if c.QueryParam("limit") != "" {
			limit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "limit query parameter must be integer")
			}
			query += fmt.Sprintf(" LIMIT %d", limit)
		}

		if err := dbConn.SelectContext(ctx, &livestreamIDs, query); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
		}
	}
	livestreamModels := getLivestreamByIDs(livestreamIDs)

	livestreams := make([]Livestream, len(livestreamModels))
	userIDs := []int64{}
	for i := range livestreamModels {
		userIDs = append(userIDs, livestreamModels[i].UserID)
	}
	userMap, err := getUserMap(ctx, dbConn, userIDs)
	if err != nil {
		return err
	}
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, dbConn, livestreamModels[i], userMap)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	return c.JSON(http.StatusOK, livestreams)
}

func getMyLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	var livestreamIDs []int64
	if err := dbConn.SelectContext(ctx, &livestreamIDs, "SELECT l.id FROM livestreams l WHERE user_id = ?", userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}
	livestreamModels := getLivestreamByIDs(livestreamIDs)
	userMap, err := getUserMap(ctx, dbConn, []int64{userID})
	if err != nil {
		return err
	}
	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, dbConn, livestreamModels[i], userMap)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	return c.JSON(http.StatusOK, livestreams)
}

func getUserLivestreamsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		return err
	}

	username := c.Param("username")

	user, exists := getUserByName(username)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	var livestreamIDs []int64
	if err := dbConn.SelectContext(ctx, &livestreamIDs, "SELECT id FROM livestreams l WHERE user_id = ?", user.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}
	livestreamModels := getLivestreamByIDs(livestreamIDs)
	userMap := map[int64]UserModel{
		user.ID: user,
	}
	livestreams := make([]Livestream, len(livestreamModels))
	for i := range livestreamModels {
		livestream, err := fillLivestreamResponse(ctx, dbConn, livestreamModels[i], userMap)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
		}
		livestreams[i] = livestream
	}

	return c.JSON(http.StatusOK, livestreams)
}

// viewerテーブルの廃止
func enterLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id must be integer")
	}

	viewer := LivestreamViewerModel{
		UserID:       int64(userID),
		LivestreamID: int64(livestreamID),
		CreatedAt:    time.Now().Unix(),
	}

	if _, err := dbConn.NamedExecContext(ctx, "INSERT INTO livestream_viewers_history (user_id, livestream_id, created_at) VALUES(:user_id, :livestream_id, :created_at)", viewer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert livestream_view_history: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func exitLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if _, err := dbConn.ExecContext(ctx, "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?", userID, livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete livestream_view_history: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func getLivestreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livestreamModel, exists := getLivestreamByID(int64(livestreamID))
	if !exists {
		return fmt.Errorf("livestream %d not found", livestreamID)
	}

	userMap, err := getUserMap(ctx, dbConn, []int64{livestreamModel.UserID})
	if err != nil {
		return err
	}
	livestream, err := fillLivestreamResponse(ctx, dbConn, livestreamModel, userMap)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livestream: "+err.Error())
	}

	return c.JSON(http.StatusOK, livestream)
}

func getLivecommentReportsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	livestreamID, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	livestreamModel, exists := getLivestreamByID(int64(livestreamID))
	if !exists {
		return fmt.Errorf("livestream %d not found", livestreamID)
	}

	sess := getSession(c)
	userID := sess.Values.UserID

	if livestreamModel.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "can't get other streamer's livecomment reports")
	}

	var reportModels []*LivecommentReportModel
	if err := dbConn.SelectContext(ctx, &reportModels, "SELECT * FROM livecomment_reports WHERE livestream_id = ?", livestreamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livecomment reports: "+err.Error())
	}

	reports := make([]LivecommentReport, len(reportModels))
	for _, r := range reportModels {
		report, err := fillLivecommentReportResponse(ctx, dbConn, *r)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to fill livecomment report: "+err.Error())
		}
		reports = append(reports, report)
	}

	return c.JSON(http.StatusOK, reports)
}

func fillLivestreamResponse(ctx context.Context, tx dbtx, livestreamModel LivestreamModel, usermap map[int64]UserModel) (Livestream, error) {
	var owner User
	if um, ok := usermap[livestreamModel.UserID]; ok {
		owner = um.toUser()
	} else {
		ownerModel := UserModel{}
		if err := tx.GetContext(ctx, &ownerModel, "SELECT * FROM users WHERE id = ?", livestreamModel.UserID); err != nil {
			return Livestream{}, err
		}
		var err error
		owner, err = fillUserResponse(ctx, tx, ownerModel)
		if err != nil {
			return Livestream{}, err
		}
	}

	rawTags := strings.Split(livestreamModel.RawTags, ",")
	tags := []Tag{}
	for _, tid := range rawTags {
		if tid != "" {
			tidint, _ := strconv.ParseInt(tid, 10, 64)
			tags = append(tags, TagMap[tidint])
		}
	}

	/*
		var livestreamTagModels []*LivestreamTagModel
		if err := tx.SelectContext(ctx, &livestreamTagModels, "SELECT * FROM livestream_tags WHERE livestream_id = ?", livestreamModel.ID); err != nil {
			return Livestream{}, err
		}

		tags := make([]Tag, len(livestreamTagModels))
		for i := range livestreamTagModels {
			tagModel := TagModel{}
			if err := tx.GetContext(ctx, &tagModel, "SELECT * FROM tags WHERE id = ?", livestreamTagModels[i].TagID); err != nil {
				return Livestream{}, err
			}

			tags[i] = Tag{
				ID:   tagModel.ID,
				Name: tagModel.Name,
			}
		}
	*/

	livestream := Livestream{
		ID:           livestreamModel.ID,
		Owner:        owner,
		Title:        livestreamModel.Title,
		Tags:         tags,
		Description:  livestreamModel.Description,
		PlaylistUrl:  livestreamModel.PlaylistUrl,
		ThumbnailUrl: livestreamModel.ThumbnailUrl,
		StartAt:      livestreamModel.StartAt,
		EndAt:        livestreamModel.EndAt,
	}
	return livestream, nil
}

func getLivestreamByID(id int64) (LivestreamModel, bool) {
	livestreamLock.RLock()
	defer livestreamLock.RUnlock()
	var livestream LivestreamModel
	var ok bool
	if livestream, ok = livestreamCache[id]; !ok {
		return LivestreamModel{}, false
	}
	return livestream, true
}

func getLivestreamByIDs(ids []int64) []LivestreamModel {
	livestreamLock.RLock()
	defer livestreamLock.RUnlock()
	livestreams := make([]LivestreamModel, len(ids))
	for i, l := range ids {
		livestreams[i] = livestreamCache[l]
	}
	return livestreams
}

func getLivestreamAll() []LivestreamModel {
	livestreamLock.RLock()
	defer livestreamLock.RUnlock()
	livestreams := make([]LivestreamModel, len(livestreamCache))
	for _, l := range livestreamCache {
		livestreams = append(livestreams, l)
	}
	return livestreams
}

func warmupLivestreamCache(ctx context.Context) {
	lCache := map[int64]LivestreamModel{}
	livestreams := []LivestreamModel{}
	query := `SELECT * FROM livestreams`
	if err := dbConn.SelectContext(ctx, &livestreams, query); err != nil {
		return
	}
	for _, l := range livestreams {
		lCache[l.ID] = l
	}
	livestreamLock.Lock()
	livestreamCache = lCache
	livestreamLock.Unlock()
}
