package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

const (
	luluWeddingPreviewLevel  = 60
	luluWeddingBlessingLevel = 68
	luluWeddingLevel         = 70
)

type luluWeddingChapter struct {
	Level       int    `json:"level"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var luluWeddingChapters = [...]luluWeddingChapter{
	{luluWeddingPreviewLevel, "抽屉里的邀请信", "噜妹给你留了一封手写信：『等噜噜到 70 级，我们想请你来见证一件小事。』"},
	{65, "摆好两只杯子", "噜噜把橘色杯子和粉色杯子摆上小桌。以后每天早起，都要给对方倒一杯热乎的。"},
	{66, "挑一束小花", "噜妹选了奶白小花，噜噜在中间悄悄加了一朵橘色的。花束已经扎好啦。"},
	{67, "把誓词装进信封", "他们各写了一句最想告诉对方的话，约好等到婚礼那天才拆开。"},
	{luluWeddingBlessingLevel, "祝福册翻到第一页", "桌上多了一本祝福册。来过这里的朋友，可以留一句想对他们说的话。"},
	{69, "只差最后一级", "灯串已经亮了，两只杯子也擦干净了。噜噜说：『再升一级，你就来帮我们数三声吧！』"},
	{luluWeddingLevel, "两只杯子的新约定", "噜噜和噜妹结婚啦。两只杯子碰在一起，今天的合照和大家的祝福都会留在回忆册里。"},
}

type luluWeddingBlessing struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

type luluWeddingVow struct {
	ID     string `json:"id"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

var luluWeddingVows = []luluWeddingVow{
	{"lulu", "噜噜", "噜妹，以后那只粉色杯子，就放在我的橘色杯子旁边。早上我先给你倒茶，晚上我们一起把杯子洗干净。你开心的时候，我陪你笑；你累的时候，我给你留着灯。"},
	{"lumei", "噜妹", "噜噜，以前我把想说的话写在信里。以后下雪也好，出太阳也好，我都想坐在你旁边，把今天的小事慢慢讲给你听。也轮到我给你泡茶啦，记得把橘色杯子递给我。"},
}

type luluWeddingState struct {
	CurrentLevel  int                   `json:"currentLevel"`
	TargetLevel   int                   `json:"targetLevel"`
	NextLevel     int                   `json:"nextLevel"`
	Chapters      []luluWeddingChapter  `json:"chapters"`
	Married       bool                  `json:"married"`
	MarriedAt     string                `json:"marriedAt"`
	CanBless      bool                  `json:"canBless"`
	BlessedToday  bool                  `json:"blessedToday"`
	BlessingCount int64                 `json:"blessingCount"`
	Blessings     []luluWeddingBlessing `json:"blessings"`
	Vows          []luluWeddingVow      `json:"vows,omitempty"`
}

func luluWeddingProgress(level int) luluWeddingState {
	state := luluWeddingState{
		CurrentLevel: level, TargetLevel: luluWeddingLevel, Married: level >= luluWeddingLevel,
		CanBless: level >= luluWeddingBlessingLevel, Chapters: []luluWeddingChapter{}, Blessings: []luluWeddingBlessing{},
	}
	for _, chapter := range luluWeddingChapters {
		if level < chapter.Level {
			state.NextLevel = chapter.Level
			break
		}
		state.Chapters = append(state.Chapters, chapter)
	}
	if state.Married {
		state.Vows = luluWeddingVows
	}
	return state
}

func (s *Service) luluWeddingLevel(ctx context.Context) (int, error) {
	rows, err := s.Repo.Query(ctx, `SELECT LEVEL FROM z_pet_status WHERE USER_NUM = ? LIMIT 1`, publicLuluUserNum)
	if err != nil || len(rows) == 0 {
		return 1, err
	}
	return max(1, model.IntValue(rows[0], "LEVEL")), nil
}

func (s *Service) ensureLuluWedding(ctx context.Context, now time.Time) (string, error) {
	const query = `SELECT MARRIED_AT FROM z_lulu_wedding WHERE USER_NUM = ?`
	rows, err := s.Repo.Query(ctx, query, publicLuluUserNum)
	if err != nil {
		return "", err
	}
	if len(rows) > 0 {
		return parseTime(model.Lookup(rows[0], "MARRIED_AT")).Format("2006-01-02"), nil
	}
	tx, err := s.Repo.DB().BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT IGNORE INTO z_lulu_wedding (USER_NUM, MARRIED_AT) VALUES (?, ?)`, publicLuluUserNum, now)
	if err != nil {
		return "", err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if inserted == 1 {
		// Save the milestone and ordinary care-log entry together, exactly once.
		_, err = tx.ExecContext(ctx, `INSERT INTO z_lulu_log
(USER_NUM, ACTION_TYPE, ACTION_NAME, IP_ADDRESS, BROWSER, DEVICE_MODEL, USER_AGENT, REMARK)
VALUES (?, 'NPC_WEDDING', '噜噜和噜妹结婚啦', 'system', '噜噜世界', '全站事件', '', 'Lv.70 婚礼：两只杯子的约定已写入回忆册。')`, publicLuluUserNum)
		if err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	rows, err = s.Repo.Query(ctx, query, publicLuluUserNum)
	if err != nil || len(rows) == 0 {
		return "", err
	}
	return parseTime(model.Lookup(rows[0], "MARRIED_AT")).Format("2006-01-02"), nil
}

func (s *Service) weddingVisitorHash(ip string) string {
	mac := hmac.New(sha256.New, []byte(s.Config.Security.JWTSecret))
	mac.Write([]byte("lulu-wedding:" + normalizeLuluIP(ip)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Service) GetLuluWedding(ctx context.Context, ip string) (luluWeddingState, error) {
	level, err := s.luluWeddingLevel(ctx)
	state := luluWeddingProgress(level)
	if err != nil || !state.CanBless {
		return state, err
	}
	now := time.Now()
	if state.Married {
		state.MarriedAt, err = s.ensureLuluWedding(ctx, now)
		if err != nil {
			return state, err
		}
	}
	counts, err := s.Repo.Query(ctx, `SELECT COUNT(*) AS TOTAL, MAX(CASE WHEN VISITOR_HASH = ? AND BLESSING_DAY = ? THEN 1 ELSE 0 END) AS BLESSED_TODAY
FROM z_lulu_wedding_blessing WHERE USER_NUM = ?`, s.weddingVisitorHash(ip), dateString(now), publicLuluUserNum)
	if err != nil {
		return state, err
	}
	state.BlessingCount = model.Int64Value(counts[0], "TOTAL")
	state.BlessedToday = model.IntValue(counts[0], "BLESSED_TODAY") == 1
	rows, err := s.Repo.Query(ctx, `SELECT ID, CONTENT FROM z_lulu_wedding_blessing WHERE USER_NUM = ? ORDER BY ID DESC LIMIT 8`, publicLuluUserNum)
	if err != nil {
		return state, err
	}
	for _, row := range rows {
		state.Blessings = append(state.Blessings, luluWeddingBlessing{model.Int64Value(row, "ID"), model.StringValue(row, "CONTENT")})
	}
	return state, nil
}

func (s *Service) AddLuluWeddingBlessing(c *gin.Context, content string) model.Result {
	ctx := c.Request.Context()
	content = strings.TrimSpace(content)
	if n := utf8.RuneCountInString(content); n < 1 || n > 60 {
		return model.Failure("请写 1 至 60 个字的祝福")
	}
	level, err := s.luluWeddingLevel(ctx)
	if err != nil {
		return dbFailure("读取婚礼进度", err)
	}
	if level < luluWeddingBlessingLevel {
		return model.Failure("祝福册将在 Lv.68 开放")
	}
	ip := ClientIP(c)
	if failure := s.checkCommentRate(ctx, ip); failure != nil {
		return *failure
	}
	subject := commentSubject{IP: ip}
	if actor, ok := s.CurrentActor(c); ok {
		subject.UserCode, subject.Nickname = actor.Code(), actor.User.Name
	}
	if failure := s.checkCommentBan(c, subject); failure != nil {
		return *failure
	}
	if s.containsBannedWord(content) {
		return model.Failure(msgBannedWord)
	}
	now := time.Now()
	affected, err := s.Repo.Exec(ctx, `INSERT IGNORE INTO z_lulu_wedding_blessing
(USER_NUM, VISITOR_HASH, BLESSING_DAY, CONTENT, CREATED_AT) VALUES (?, ?, ?, ?, ?)`, publicLuluUserNum, s.weddingVisitorHash(ip), dateString(now), content, now)
	if err != nil {
		return dbFailure("保存祝福", err)
	}
	if affected != 1 {
		return model.Failure("今天已经送过祝福啦")
	}
	s.countComment(ctx, ip)
	return model.Success(nil)
}
