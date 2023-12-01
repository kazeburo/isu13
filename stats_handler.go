package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type LivestreamStatistics struct {
	Rank           int64 `json:"rank"`
	ViewersCount   int64 `json:"viewers_count"`
	TotalReactions int64 `json:"total_reactions"`
	TotalReports   int64 `json:"total_reports"`
	MaxTip         int64 `json:"max_tip"`
}

type LivestreamRankingEntry struct {
	LivestreamID int64
	Score        int64
}
type LivestreamRanking []LivestreamRankingEntry

func (r LivestreamRanking) Len() int      { return len(r) }
func (r LivestreamRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r LivestreamRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].LivestreamID < r[j].LivestreamID
	} else {
		return r[i].Score < r[j].Score
	}
}

type UserStatistics struct {
	Rank              int64  `json:"rank"`
	ViewersCount      int64  `json:"viewers_count"`
	TotalReactions    int64  `json:"total_reactions"`
	TotalLivecomments int64  `json:"total_livecomments"`
	TotalTip          int64  `json:"total_tip"`
	FavoriteEmoji     string `json:"favorite_emoji"`
}

type UserRankingEntry struct {
	Username string `db:"username"`
	Score    int64
	Reaction int64 `db:"reaction"`
	Tips     int64 `db:"tips"`
}
type UserRanking []UserRankingEntry

func (r UserRanking) Len() int      { return len(r) }
func (r UserRanking) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r UserRanking) Less(i, j int) bool {
	if r[i].Score == r[j].Score {
		return r[i].Username < r[j].Username
	} else {
		return r[i].Score < r[j].Score
	}
}

func getUserStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	username := c.Param("username")
	// ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
	// また、現在の合計視聴者数もだす

	user, exists := getUserByName(username)
	if !exists {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}

	// ランク算出
	query := `
	WITH reaction_per_user AS (
    SELECT l.user_id, COUNT(*) AS reaction_count FROM reactions r
    LEFT JOIN livestreams l ON l.id = r.livestream_id GROUP BY l.user_id
), tip_per_user AS (
    SELECT l.user_id, IFNULL(SUM(lc.tip), 0) AS sum_tip FROM livecomments lc
    LEFT JOIN livestreams l ON l.id = lc.livestream_id GROUP BY l.user_id
), ranking_score AS (
    SELECT reaction_per_user.user_id, (IFNULL(reaction_count, 0) + IFNULL(sum_tip, 0)) AS score FROM reaction_per_user LEFT OUTER JOIN tip_per_user ON reaction_per_user.user_id = tip_per_user.user_id
), ranking_per_user AS (
    SELECT users.id AS user_id, IFNULL(ranking_score.score, 0), ROW_NUMBER() OVER w AS 'ranking' FROM users LEFT JOIN ranking_score ON users.id = ranking_score.user_id WINDOW w AS (ORDER BY ranking_score.score DESC, users.name DESC)
)
SELECT ranking FROM ranking_per_user WHERE user_id = ?
`
	var rank int64
	if err := dbConn.GetContext(ctx, &rank, query, user.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count rank reactions: "+err.Error())
	}

	// リアクション数
	var totalReactions int64
	query = `SELECT COUNT(*) FROM users u 
    INNER JOIN livestreams l ON l.user_id = u.id 
    INNER JOIN reactions r ON r.livestream_id = l.id
    WHERE u.name = ?
	`
	if err := dbConn.GetContext(ctx, &totalReactions, query, username); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total reactions: "+err.Error())
	}

	// ライブコメント数、チップ合計
	var totalLivecomments int64
	var totalTip int64
	var livestreams []*LivestreamModel
	if err := dbConn.SelectContext(ctx, &livestreams, "SELECT * FROM livestreams WHERE user_id = ?", user.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestreams: "+err.Error())
	}

	for _, livestream := range livestreams {
		var livecomments []*LivecommentModel
		if err := dbConn.SelectContext(ctx, &livecomments, "SELECT * FROM livecomments WHERE livestream_id = ?", livestream.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livecomments: "+err.Error())
		}

		for _, livecomment := range livecomments {
			totalTip += livecomment.Tip
			totalLivecomments++
		}
	}

	// 合計視聴者数
	var viewersCount int64
	for _, livestream := range livestreams {
		var cnt int64
		if err := dbConn.GetContext(ctx, &cnt, "SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?", livestream.ID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestream_view_history: "+err.Error())
		}
		viewersCount += cnt
	}

	// お気に入り絵文字
	var favoriteEmoji string
	query = `
	SELECT r.emoji_name
	FROM users u
	INNER JOIN livestreams l ON l.user_id = u.id
	INNER JOIN reactions r ON r.livestream_id = l.id
	WHERE u.name = ?
	GROUP BY emoji_name
	ORDER BY COUNT(*) DESC, emoji_name DESC
	LIMIT 1
	`
	if err := dbConn.GetContext(ctx, &favoriteEmoji, query, username); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find favorite emoji: "+err.Error())
	}

	stats := UserStatistics{
		Rank:              rank,
		ViewersCount:      viewersCount,
		TotalReactions:    totalReactions,
		TotalLivecomments: totalLivecomments,
		TotalTip:          totalTip,
		FavoriteEmoji:     favoriteEmoji,
	}
	return c.JSON(http.StatusOK, stats)
}

func getLivestreamStatisticsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	id, err := strconv.Atoi(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}
	livestreamID := int64(id)

	/*
		var livestream LivestreamModel
		if err := dbConn.GetContext(ctx, &livestream, "SELECT * FROM livestreams WHERE id = ?", livestreamID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return echo.NewHTTPError(http.StatusBadRequest, "cannot get stats of not found livestream")
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get livestream: "+err.Error())
			}
		}
	*/

	// ランク算出
	var rank int64
	query := `
		WITH reaction_per_stream AS (
		  SELECT r.livestream_id AS livestream_id, COUNT(*) AS reaction_count FROM reactions r
		GROUP BY r.livestream_id
		), tip_per_strea AS (
		  SELECT c.livestream_id AS livestream_id, IFNULL(SUM(c.tip), 0) AS sum_tip FROM livecomments c GROUP BY c.livestream_id
		), ranking_score AS (
		  SELECT reaction_per_stream.livestream_id, (IFNULL(reaction_count, 0) + IFNULL(sum_tip, 0)) AS score FROM reaction_per_stream LEFT OUTER JOIN tip_per_strea ON reaction_per_stream.livestream_id = tip_per_strea.livestream_id
		), ranking_per_stream AS (
		  SELECT livestreams.id AS livestream_id, IFNULL(ranking_score.score, 0), ROW_NUMBER() OVER w AS 'ranking' FROM livestreams LEFT JOIN ranking_score ON livestreams.id = ranking_score.livestream_id WINDOW w AS (ORDER BY ranking_score.score DESC, livestreams.id DESC)
		)
		SELECT ranking FROM ranking_per_stream WHERE livestream_id = ?
	`
	if err := dbConn.GetContext(ctx, &rank, query, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count rank reactions: "+err.Error())
	}

	// 視聴者数算出
	var viewersCount int64
	if err := dbConn.GetContext(ctx, &viewersCount, `SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count livestream viewers: "+err.Error())
	}

	// 最大チップ額
	var maxTip int64
	if err := dbConn.GetContext(ctx, &maxTip, `SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find maximum tip livecomment: "+err.Error())
	}

	// リアクション数
	var totalReactions int64
	if err := dbConn.GetContext(ctx, &totalReactions, "SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?", livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total reactions: "+err.Error())
	}

	// スパム報告数
	var totalReports int64
	if err := dbConn.GetContext(ctx, &totalReports, `SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?`, livestreamID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count total spam reports: "+err.Error())
	}
	return c.JSON(http.StatusOK, LivestreamStatistics{
		Rank:           rank,
		ViewersCount:   viewersCount,
		MaxTip:         maxTip,
		TotalReactions: totalReactions,
		TotalReports:   totalReports,
	})
}
