package service

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"nys-go-api/internal/model"
)

// 违禁词 = config.yaml 中的 content.banned_words（只读）+ 数据库 bannedWord 表（后台在线增删）。
// 数据库部分在内存中缓存 60 秒，后台修改后立即刷新；表不存在（未执行迁移 007）时只使用配置文件中的词。

const bannedWordCacheTTL = 60 * time.Second

// bannedWordMaxLength 与 bannedWord.WORD 的列宽一致
const bannedWordMaxLength = 100

var bannedWordCache struct {
	sync.Mutex
	words    []string
	loadedAt time.Time
}

func (s *Service) customBannedWords(ctx context.Context) []string {
	bannedWordCache.Lock()
	defer bannedWordCache.Unlock()
	if time.Since(bannedWordCache.loadedAt) < bannedWordCacheTTL {
		return bannedWordCache.words
	}
	if s.Repo == nil {
		return nil
	}
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := s.Repo.Query(queryContext, "SELECT WORD FROM bannedWord")
	// 无论成功与否都记下时间，表不存在时不会每条评论都去查一次数据库
	bannedWordCache.loadedAt = time.Now()
	if err != nil {
		bannedWordCache.words = nil
		return nil
	}
	words := make([]string, 0, len(rows))
	for _, row := range rows {
		if word := strings.TrimSpace(model.StringValue(row, "WORD")); word != "" {
			words = append(words, word)
		}
	}
	bannedWordCache.words = words
	return words
}

func invalidateBannedWordCache() {
	bannedWordCache.Lock()
	bannedWordCache.loadedAt = time.Time{}
	bannedWordCache.Unlock()
}

func (s *Service) containsBannedWord(content string) bool {
	content = strings.ToLower(strings.TrimSpace(content))
	if content == "" {
		return false
	}
	words := append(append([]string{}, s.Config.Content.BannedWords...), s.customBannedWords(context.Background())...)
	for _, word := range words {
		if word != "" && strings.Contains(content, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// GetBannedWords 返回配置文件中的词（只读）和后台添加的词。
func (s *Service) GetBannedWords(c *gin.Context) model.Result {
	if _, failure := s.requireSuperAdmin(c); failure != nil {
		return *failure
	}
	rows, err := s.Repo.Query(contextOf(c), "SELECT WORD, CREATE_USER, CREATE_TIME FROM bannedWord ORDER BY CREATE_TIME DESC")
	tableReady := err == nil
	if err != nil {
		log.Printf("读取违禁词表失败（是否已执行迁移 007？）: %v", err)
		rows = []map[string]any{}
	}
	return model.Success(map[string]any{
		"config":     s.Config.Content.BannedWords,
		"custom":     rows,
		"tableReady": tableReady,
	})
}

func (s *Service) AddBannedWord(c *gin.Context, word string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	word = strings.TrimSpace(word)
	if word == "" {
		return model.Failure("违禁词不能为空")
	}
	if len([]rune(word)) > bannedWordMaxLength {
		return model.Failure("违禁词不能超过 100 个字")
	}
	if _, err := s.Repo.Exec(contextOf(c), "INSERT IGNORE INTO bannedWord (WORD, CREATE_USER) VALUES (?, ?)", word, actor.Code()); err != nil {
		return model.Failure("添加失败，请确认已执行迁移 007_add_admin_tables.sql: " + err.Error())
	}
	invalidateBannedWordCache()
	s.recordAdminAction(c, actor, AdminActionBannedWord, word, "添加违禁词")
	return model.Success("已添加")
}

func (s *Service) DeleteBannedWord(c *gin.Context, word string) model.Result {
	actor, failure := s.requireSuperAdmin(c)
	if failure != nil {
		return *failure
	}
	word = strings.TrimSpace(word)
	if _, err := s.Repo.Exec(contextOf(c), "DELETE FROM bannedWord WHERE WORD = ?", word); err != nil {
		return dbFailure("删除违禁词", err)
	}
	invalidateBannedWordCache()
	s.recordAdminAction(c, actor, AdminActionBannedWord, word, "删除违禁词")
	return model.Success("已删除")
}
