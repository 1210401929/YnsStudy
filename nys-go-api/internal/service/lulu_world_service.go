package service

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"nys-go-api/internal/model"
)

const (
	luluNPCEventRollMax      = 1000
	luluNPCOutingRollLimit   = 500
	luluNPCLetterRollLimit   = 800
	luluNPCEventCooldownDays = 1
)

type luluNPCEventSpec struct {
	EventType string
	Title     string
	Content   string
	Duration  time.Duration
}

type luluCommunityGoalSpec struct {
	GoalType  string
	Title     string
	Target    int
	Condition string
	Reward    string
}

var lumeiLetterContents = []string{
	"我在路边看见一朵像小太阳的花，第一时间就想送给你。今天也要开心呀！",
	"今天的云像一大块软软的棉花糖。替我告诉噜噜，下次要一起去看。",
	"我偷偷准备了新的游戏，下次见面要和噜噜一决胜负。先帮我保密哦！",
	"路过小面包店时闻到了橘子香，我猜噜噜一定会喜欢，所以写信来提醒它记得吃饭。",
}

var luluMemorySignatures = []string{
	"噜噜把你的脚步声认真记进了回忆册。",
	"每次你来，这间小屋好像都会亮一点。",
	"你留下的陪伴，噜噜一件也没有忘记。",
	"噜噜已经能认出这位常来坐坐的朋友啦。",
}

func (s *Service) GetPetWorld(ctx context.Context, userNum int64, ip, userAgent string) (map[string]any, error) {
	ip = normalizeLuluIP(ip)
	now := time.Now()
	memory, visitDays, err := s.getLuluMemory(ctx, userNum, ip, now)
	if err != nil {
		return nil, err
	}
	event, history, err := s.resolveLuluNPCEvent(ctx, userNum, ip, userAgent, visitDays, now)
	if err != nil {
		return nil, err
	}
	goal, err := s.getLuluCommunityGoal(ctx, userNum, now)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"NPC_EVENT":      event,
		"NPC_HISTORY":    history,
		"MEMORY":         memory,
		"COMMUNITY_GOAL": goal,
	}, nil
}

func (s *Service) getLuluMemory(ctx context.Context, userNum int64, ip string, now time.Time) (map[string]any, int, error) {
	rows, err := s.Repo.Query(ctx, `SELECT COUNT(DISTINCT DATE(CREATE_TIME)) AS VISIT_DAYS,
MIN(CREATE_TIME) AS FIRST_VISIT_TIME,
SUM(CASE WHEN ACTION_TYPE IN ('FEED', 'PLAY', 'AUTO_CARE', 'SLEEP', 'WAKE') THEN 1 ELSE 0 END) AS INTERACTION_COUNT
FROM z_lulu_log WHERE USER_NUM = ? AND IP_ADDRESS = ?`, userNum, ip)
	if err != nil {
		return nil, 0, err
	}
	stats := map[string]any{}
	if len(rows) > 0 {
		stats = rows[0]
	}
	visitDays := model.IntValue(stats, "VISIT_DAYS")
	interactionCount := model.IntValue(stats, "INTERACTION_COUNT")
	firstVisit := parseTime(model.Lookup(stats, "FIRST_VISIT_TIME"))

	favoriteRows, err := s.Repo.Query(ctx, `SELECT ACTION_NAME, COUNT(1) AS TOTAL
FROM z_lulu_log WHERE USER_NUM = ? AND IP_ADDRESS = ?
AND ACTION_TYPE IN ('FEED', 'PLAY', 'AUTO_CARE', 'SLEEP', 'WAKE')
GROUP BY ACTION_NAME ORDER BY TOTAL DESC, ACTION_NAME ASC LIMIT 1`, userNum, ip)
	if err != nil {
		return nil, 0, err
	}
	favoriteAction := "还在慢慢认识噜噜"
	if len(favoriteRows) > 0 && model.StringValue(favoriteRows[0], "ACTION_NAME") != "" {
		favoriteAction = model.StringValue(favoriteRows[0], "ACTION_NAME")
	}

	messageRows, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM z_lulu_message WHERE USER_NUM = ? AND IP_ADDRESS = ?", userNum, ip)
	if err != nil {
		return nil, 0, err
	}
	messageCount := int(firstCount(messageRows))
	return luluMemoryResponse(now, ip, visitDays, interactionCount, messageCount, firstVisit, favoriteAction), visitDays, nil
}

func luluMemoryResponse(now time.Time, ip string, visitDays, interactionCount, messageCount int, firstVisit time.Time, favoriteAction string) map[string]any {
	if visitDays < 1 {
		visitDays = 1
		firstVisit = now
	}
	daysKnown := maxInt(1, int(now.Sub(firstVisit).Hours()/24)+1)
	lines := []string{
		"第一次见面是 " + firstVisit.Format("2006年1月2日") + "，已经认识 " + intText(daysKnown) + " 天。",
		"你有 " + intText(visitDays) + " 天来陪过噜噜，一共完成了 " + intText(interactionCount) + " 次照顾。",
		"你最常做的是「" + favoriteAction + "」，还留下了 " + intText(messageCount) + " 张小纸条。",
	}
	return map[string]any{
		"TITLE":             luluMemoryTitle(visitDays),
		"SIGNATURE":         stableLuluMemorySignature(ip),
		"LINES":             lines,
		"FIRST_VISIT_DATE":  firstVisit.Format("2006-01-02"),
		"DAYS_KNOWN":        daysKnown,
		"VISIT_DAYS":        visitDays,
		"INTERACTION_COUNT": interactionCount,
		"MESSAGE_COUNT":     messageCount,
		"FAVORITE_ACTION":   favoriteAction,
	}
}

func luluMemoryTitle(visitDays int) string {
	switch {
	case visitDays >= 60:
		return "噜噜最珍惜的老朋友"
	case visitDays >= 30:
		return "不会忘记的熟悉身影"
	case visitDays >= 7:
		return "噜噜认真记住的人"
	case visitDays >= 2:
		return "常来坐坐的小伙伴"
	default:
		return "今天认识的新朋友"
	}
}

func stableLuluMemorySignature(ip string) string {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(ip))
	return luluMemorySignatures[int(hasher.Sum32())%len(luluMemorySignatures)]
}

func stableLuluEventRoll(ip, eventDate string) int {
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(ip + "|" + eventDate + "|lulu-npc"))
	return int(hasher.Sum32() % luluNPCEventRollMax)
}

func (s *Service) resolveLuluNPCEvent(ctx context.Context, userNum int64, ip, userAgent string, visitDays int, now time.Time) (any, []map[string]any, error) {
	today := dateString(now)
	todayRows, err := s.Repo.Query(ctx, `SELECT * FROM z_lulu_npc_event
WHERE USER_NUM = ? AND IP_ADDRESS = ? AND EVENT_DATE = ? ORDER BY ID DESC LIMIT 1`, userNum, ip, today)
	if err != nil {
		return nil, nil, err
	}

	if len(todayRows) == 0 {
		lastRows, queryErr := s.Repo.Query(ctx, `SELECT * FROM z_lulu_npc_event
WHERE USER_NUM = ? AND IP_ADDRESS = ? ORDER BY EVENT_DATE DESC, ID DESC LIMIT 1`, userNum, ip)
		if queryErr != nil {
			return nil, nil, queryErr
		}
		daysSinceLast := 9999
		if len(lastRows) > 0 {
			lastDate := parseTime(model.Lookup(lastRows[0], "EVENT_DATE"))
			daysSinceLast = int(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Sub(lastDate).Hours() / 24)
		}
		// Keep the result stable for the whole IP/day. Refreshing the page cannot
		// repeatedly reroll the easter egg and turn a rare event into a common one.
		spec := chooseLuluNPCEvent(visitDays, daysSinceLast, stableLuluEventRoll(ip, today))
		if spec != nil {
			expiresAt := now.Add(spec.Duration)
			affected, insertErr := s.Repo.Exec(ctx, `INSERT IGNORE INTO z_lulu_npc_event
(USER_NUM, IP_ADDRESS, EVENT_DATE, EVENT_TYPE, TITLE, CONTENT, EXPIRES_AT)
VALUES (?, ?, ?, ?, ?, ?, ?)`, userNum, ip, today, spec.EventType, spec.Title, spec.Content, expiresAt)
			if insertErr != nil {
				return nil, nil, insertErr
			}
			if affected > 0 {
				actionType, actionName := "NPC_LETTER", "噜妹来信"
				if spec.EventType == "OUTING" {
					actionType, actionName = "NPC_OUTING", "和噜妹外出"
				}
				s.insertPetLog(ctx, userNum, actionType, actionName, ip, userAgent, spec.Title+"："+spec.Content)
			}
			todayRows, err = s.Repo.Query(ctx, `SELECT * FROM z_lulu_npc_event
WHERE USER_NUM = ? AND IP_ADDRESS = ? AND EVENT_DATE = ? ORDER BY ID DESC LIMIT 1`, userNum, ip, today)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	var currentEvent any
	if len(todayRows) > 0 {
		currentEvent = luluNPCEventResponse(todayRows[0], now)
	}
	historyRows, err := s.Repo.Query(ctx, `SELECT * FROM z_lulu_npc_event
WHERE USER_NUM = ? AND IP_ADDRESS = ? ORDER BY EVENT_DATE DESC, ID DESC LIMIT 8`, userNum, ip)
	if err != nil {
		return nil, nil, err
	}
	history := make([]map[string]any, 0, len(historyRows))
	for _, row := range historyRows {
		history = append(history, luluNPCEventResponse(row, now))
	}
	return currentEvent, history, nil
}

func chooseLuluNPCEvent(visitDays, daysSinceLast, roll int) *luluNPCEventSpec {
	if visitDays < 2 || daysSinceLast < luluNPCEventCooldownDays || roll < 0 || roll >= luluNPCEventRollMax {
		return nil
	}
	if visitDays >= 5 && roll < luluNPCOutingRollLimit {
		return &luluNPCEventSpec{
			EventType: "OUTING",
			Title:     "他们悄悄出门啦",
			Content:   "噜噜和噜妹背着小包去公园找云朵形状了，大约十分钟后就会回来。",
			Duration:  10 * time.Minute,
		}
	}
	if roll < luluNPCLetterRollLimit {
		return &luluNPCEventSpec{
			EventType: "LETTER",
			Title:     "噜妹寄来一封信",
			Content:   lumeiLetterContents[roll%len(lumeiLetterContents)],
			Duration:  6 * time.Hour,
		}
	}
	return nil
}

func luluNPCEventResponse(row map[string]any, now time.Time) map[string]any {
	expiresAt := parseTime(model.Lookup(row, "EXPIRES_AT"))
	return map[string]any{
		"ID":          model.Int64Value(row, "ID"),
		"EVENT_TYPE":  model.StringValue(row, "EVENT_TYPE"),
		"EVENT_DATE":  model.StringValue(row, "EVENT_DATE"),
		"TITLE":       model.StringValue(row, "TITLE"),
		"CONTENT":     model.StringValue(row, "CONTENT"),
		"EXPIRES_AT":  expiresAt.Format(time.RFC3339),
		"ACTIVE":      expiresAt.After(now),
		"CREATE_TIME": model.StringValue(row, "CREATE_TIME"),
	}
}

func (s *Service) getLuluCommunityGoal(ctx context.Context, userNum int64, now time.Time) (map[string]any, error) {
	spec := luluCommunityGoalForDate(now)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextDay := dayStart.AddDate(0, 0, 1)
	query := `SELECT COUNT(1) AS TOTAL, COUNT(DISTINCT NULLIF(IP_ADDRESS, '')) AS PARTICIPANTS
FROM z_lulu_log WHERE USER_NUM = ? AND CREATE_TIME >= ? AND CREATE_TIME < ? AND ` + spec.Condition
	rows, err := s.Repo.Query(ctx, query, userNum, dayStart, nextDay)
	if err != nil {
		return nil, err
	}
	current, participants := 0, 0
	if len(rows) > 0 {
		current = model.IntValue(rows[0], "TOTAL")
		participants = model.IntValue(rows[0], "PARTICIPANTS")
	}
	return map[string]any{
		"GOAL_DATE":    dateString(now),
		"GOAL_TYPE":    spec.GoalType,
		"TITLE":        spec.Title,
		"TARGET":       spec.Target,
		"CURRENT":      current,
		"REMAINING":    maxInt(0, spec.Target-current),
		"PARTICIPANTS": participants,
		"COMPLETED":    current >= spec.Target,
		"REWARD":       spec.Reward,
	}, nil
}

func luluCommunityGoalForDate(now time.Time) luluCommunityGoalSpec {
	goals := []luluCommunityGoalSpec{
		{GoalType: "FEED", Title: "大家一起给噜噜准备 20 顿饭", Target: 20, Condition: "ACTION_TYPE = 'FEED'", Reward: "全站解锁一整天的橘子香心情"},
		{GoalType: "PLAY", Title: "大家一起陪噜噜玩耍 20 次", Target: 20, Condition: "ACTION_TYPE = 'PLAY'", Reward: "全站解锁一整天的闪亮心情"},
		{GoalType: "CARE", Title: "大家共同完成 30 次照顾", Target: 30, Condition: "ACTION_TYPE IN ('FEED', 'PLAY', 'AUTO_CARE', 'SLEEP', 'WAKE')", Reward: "全站解锁一整天的温暖心情"},
	}
	return goals[now.YearDay()%len(goals)]
}

func normalizeLuluIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "unknown"
	}
	return truncateRunes(ip, 64)
}

func intText(value int) string {
	if value < 0 {
		value = 0
	}
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	buffer := make([]byte, 0, 12)
	for value > 0 {
		buffer = append(buffer, digits[value%10])
		value /= 10
	}
	for left, right := 0, len(buffer)-1; left < right; left, right = left+1, right-1 {
		buffer[left], buffer[right] = buffer[right], buffer[left]
	}
	return string(buffer)
}
