package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"nys-go-api/internal/model"
)

const defaultPetName = "噜噜"

var dailyMissions = [][4]string{
	{"feed", "给噜噜准备一顿饭", "1", "完成后心情会亮一下"},
	{"play", "陪噜噜玩两次", "2", "完成后撒一把星星"},
	{"touch", "摸摸噜噜一次", "1", "完成后获得贴贴感"},
	{"wish", "和噜噜许个愿", "1", "完成后收到小签语"},
}

type smartCarePlan struct {
	Action        string
	ActionName    string
	Message       string
	MissionType   string
	HungerDelta   int
	EnergyDelta   int
	MoodDelta     int
	Experience    int
	NextState     string
	ShouldPersist bool
}

func (s *Service) GetPetStatus(ctx context.Context, userNum int64) (map[string]any, error) {
	rows, err := s.Repo.Query(ctx, "SELECT * FROM z_pet_status WHERE USER_NUM = ?", userNum)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return s.createDefaultPet(ctx, userNum)
	}
	pet := rows[0]
	if model.Lookup(pet, "LEVEL") == nil {
		pet["LEVEL"] = int64(1)
	}
	if model.Lookup(pet, "EXP") == nil {
		pet["EXP"] = int64(0)
	}
	if err := s.refreshPet(ctx, pet); err != nil {
		return nil, err
	}
	return pet, nil
}

func (s *Service) FeedPet(ctx context.Context, userNum int64, ip, userAgent string) (map[string]any, error) {
	pet, err := s.GetPetStatus(ctx, userNum)
	if err != nil {
		return nil, err
	}
	pet["HUNGER"] = minInt(100, model.IntValue(pet, "HUNGER")+30)
	pet["MOOD"] = minInt(100, model.IntValue(pet, "MOOD")+5)
	pet["CURRENT_STATE"] = "IDLE"
	pet["LAST_UPDATE_TIME"] = time.Now()
	addPetExperience(pet, 20)
	if err := s.updatePet(ctx, pet); err != nil {
		return nil, err
	}
	s.insertPetLog(ctx, userNum, "FEED", "喂食", ip, userAgent, "饱腹 +30，心情 +5，经验 +20")
	return pet, nil
}

func (s *Service) PlayPet(ctx context.Context, userNum int64, action, ip, userAgent string) (map[string]any, error) {
	pet, err := s.GetPetStatus(ctx, userNum)
	if err != nil {
		return nil, err
	}
	action = strings.TrimSpace(action)
	if action == "" {
		action = "玩耍"
	}
	if model.StringValue(pet, "CURRENT_STATE") == "SLEEPING" || model.IntValue(pet, "ENERGY") < 15 {
		return pet, nil
	}
	pet["ENERGY"] = maxInt(0, model.IntValue(pet, "ENERGY")-15)
	pet["HUNGER"] = maxInt(0, model.IntValue(pet, "HUNGER")-10)
	pet["MOOD"] = minInt(100, model.IntValue(pet, "MOOD")+20)
	pet["CURRENT_STATE"] = "IDLE"
	pet["LAST_UPDATE_TIME"] = time.Now()
	addPetExperience(pet, 40)
	if err := s.updatePet(ctx, pet); err != nil {
		return nil, err
	}
	s.insertPetLog(ctx, userNum, "PLAY", action, ip, userAgent, action+"，心情 +20，经验 +40")
	return pet, nil
}

// SmartCarePet chooses one small, state-aware action for the pet. It deliberately
// reuses z_pet_status and z_lulu_log so the feature needs no schema migration.
func (s *Service) SmartCarePet(ctx context.Context, userNum int64, ip, userAgent string) (map[string]any, error) {
	pet, err := s.GetPetStatus(ctx, userNum)
	if err != nil {
		return nil, err
	}

	plan := selectSmartCarePlan(pet)
	if plan.ShouldPersist {
		pet["HUNGER"] = clampPetStat(model.IntValue(pet, "HUNGER") + plan.HungerDelta)
		pet["ENERGY"] = clampPetStat(model.IntValue(pet, "ENERGY") + plan.EnergyDelta)
		pet["MOOD"] = clampPetStat(model.IntValue(pet, "MOOD") + plan.MoodDelta)
		if plan.NextState != "" {
			pet["CURRENT_STATE"] = plan.NextState
		}
		pet["LAST_UPDATE_TIME"] = time.Now()
		addPetExperience(pet, plan.Experience)
		if err := s.updatePet(ctx, pet); err != nil {
			return nil, err
		}
		s.insertPetLog(ctx, userNum, "AUTO_CARE", plan.ActionName, ip, userAgent, plan.Message)
	}

	pet["CARE_ACTION"] = plan.Action
	pet["CARE_MESSAGE"] = plan.Message
	pet["CARE_MISSION_TYPE"] = plan.MissionType
	pet["CARE_EXP_GAIN"] = plan.Experience
	pet["ACTION_ACCEPTED"] = plan.ShouldPersist
	return pet, nil
}

// RecordPetVisit writes at most one passive visit log per day. Active care logs
// already prove a visit, so opening the page later on the same day adds nothing.
func (s *Service) RecordPetVisit(ctx context.Context, userNum int64, ip, userAgent string) (map[string]any, error) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nextDay := dayStart.AddDate(0, 0, 1)
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM z_lulu_log WHERE USER_NUM = ? AND CREATE_TIME >= ? AND CREATE_TIME < ?", userNum, dayStart, nextDay)
	if err != nil {
		return nil, err
	}
	if firstCount(counts) == 0 {
		if strings.TrimSpace(ip) == "" {
			ip = "unknown"
		}
		_, err = s.Repo.Exec(ctx, `INSERT INTO z_lulu_log
(USER_NUM, ACTION_TYPE, ACTION_NAME, IP_ADDRESS, BROWSER, DEVICE_MODEL, USER_AGENT, REMARK)
VALUES (?, 'VISIT', '来看噜噜', ?, ?, ?, ?, '今天第一次来看噜噜')`, userNum, ip, parseBrowser(userAgent), parseDevice(userAgent), truncateRunes(userAgent, 500))
		if err != nil {
			return nil, err
		}
	}
	return s.GetMonthlyCompanionship(ctx, userNum)
}

func selectSmartCarePlan(pet map[string]any) smartCarePlan {
	if model.StringValue(pet, "CURRENT_STATE") == "SLEEPING" {
		return smartCarePlan{
			Action: "RESTING", ActionName: "守护睡眠", Message: "噜噜正在安心休息，不打扰就是最好的照顾。",
		}
	}
	if model.IntValue(pet, "HUNGER") <= 45 {
		return smartCarePlan{
			Action: "FEED", ActionName: "智能加餐", Message: "噜噜的小肚子在提醒你：先补充一点能量吧。", MissionType: "feed",
			HungerDelta: 30, MoodDelta: 5, Experience: 20, NextState: "IDLE", ShouldPersist: true,
		}
	}
	if model.IntValue(pet, "ENERGY") <= 35 {
		return smartCarePlan{
			Action: "REST", ActionName: "安排休息", Message: "噜噜有些累了，已经为它铺好小床。",
			Experience: 5, NextState: "SLEEPING", ShouldPersist: true,
		}
	}
	if model.IntValue(pet, "MOOD") <= 65 {
		return smartCarePlan{
			Action: "COMFORT", ActionName: "温柔陪伴", Message: "一个摸摸和一点陪伴，让噜噜重新开心起来。", MissionType: "touch",
			HungerDelta: -2, EnergyDelta: -3, MoodDelta: 15, Experience: 20, NextState: "IDLE", ShouldPersist: true,
		}
	}
	return smartCarePlan{
		Action: "STROLL", ActionName: "一起散步", Message: "状态正好，噜噜想和你出去走走。", MissionType: "play",
		HungerDelta: -5, EnergyDelta: -8, MoodDelta: 10, Experience: 25, NextState: "IDLE", ShouldPersist: true,
	}
}

func clampPetStat(value int) int {
	return maxInt(0, minInt(100, value))
}

func (s *Service) TogglePetSleep(ctx context.Context, userNum int64, ip, userAgent string) (map[string]any, error) {
	pet, err := s.GetPetStatus(ctx, userNum)
	if err != nil {
		return nil, err
	}
	waking := model.StringValue(pet, "CURRENT_STATE") == "SLEEPING"
	state, actionType, actionName, remark := "SLEEPING", "SLEEP", "睡觉", "让 噜噜 休息"
	if waking {
		state, actionType, actionName, remark = "IDLE", "WAKE", "唤醒", "把 噜噜 唤醒了"
	}
	pet["CURRENT_STATE"] = state
	pet["LAST_UPDATE_TIME"] = time.Now()
	if err := s.updatePet(ctx, pet); err != nil {
		return nil, err
	}
	s.insertPetLog(ctx, userNum, actionType, actionName, ip, userAgent, remark)
	return pet, nil
}

func (s *Service) GetPetMessages(ctx context.Context, userNum int64, page, pageSize int) (map[string]any, error) {
	page, pageSize = normalizeLuluPage(page, pageSize, 8)
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM z_lulu_message WHERE USER_NUM = ?", userNum)
	if err != nil {
		return nil, err
	}
	total := firstCount(counts)
	page = clampLuluPage(page, pageSize, total)
	rows, err := s.Repo.Query(ctx, "SELECT * FROM z_lulu_message WHERE USER_NUM = ? ORDER BY CREATE_TIME DESC, ID DESC LIMIT ? OFFSET ?", userNum, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return luluPageResponse(rows, total, page, pageSize), nil
}

func (s *Service) AddPetMessage(ctx context.Context, userNum int64, content, ip string) model.Result {
	content = truncateRunes(strings.TrimSpace(content), 500)
	if content == "" {
		return model.Failure("留言内容不能为空")
	}
	if strings.TrimSpace(ip) == "" {
		ip = "unknown"
	}
	rows, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS MESSAGE_COUNT FROM z_lulu_message WHERE IP_ADDRESS = ? AND CREATE_TIME >= ?", ip, time.Now().Add(-time.Duration(s.Config.Content.MessageWindowSecond)*time.Second))
	if err != nil {
		return dbFailure("检查留言频率", err)
	}
	if firstCount(rows) >= int64(s.Config.Content.MessageLimit) {
		return model.Failure("留言太频繁了，同一 IP 1 分钟内最多发布 5 条留言")
	}
	if _, err := s.Repo.Exec(ctx, "INSERT INTO z_lulu_message (USER_NUM, SENDER_TYPE, CONTENT, IP_ADDRESS) VALUES (?, 'USER', ?, ?)", userNum, content, ip); err != nil {
		return dbFailure("新增留言", err)
	}
	return model.Success(map[string]any{"CREATED": true})
}

func (s *Service) DeletePetMessage(ctx context.Context, userNum, messageID int64) (map[string]any, error) {
	affected, err := s.Repo.Exec(ctx, "DELETE FROM z_lulu_message WHERE ID = ? AND USER_NUM = ?", messageID, userNum)
	if err != nil {
		return nil, err
	}
	return map[string]any{"DELETED": affected > 0}, nil
}

func (s *Service) GetPetLogs(ctx context.Context, userNum int64, page, pageSize int) (map[string]any, error) {
	page, pageSize = normalizeLuluPage(page, pageSize, 10)
	counts, err := s.Repo.Query(ctx, "SELECT COUNT(1) AS TOTAL FROM z_lulu_log WHERE USER_NUM = ?", userNum)
	if err != nil {
		return nil, err
	}
	total := firstCount(counts)
	page = clampLuluPage(page, pageSize, total)
	rows, err := s.Repo.Query(ctx, "SELECT * FROM z_lulu_log WHERE USER_NUM = ? ORDER BY CREATE_TIME DESC, ID DESC LIMIT ? OFFSET ?", userNum, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return luluPageResponse(rows, total, page, pageSize), nil
}

func (s *Service) GetMonthlyCompanionship(ctx context.Context, userNum int64) (map[string]any, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)
	rows, err := s.Repo.Query(ctx, `SELECT DISTINCT DATE_FORMAT(CREATE_TIME, '%Y-%m-%d') AS VISIT_DATE
FROM z_lulu_log WHERE USER_NUM = ? AND CREATE_TIME >= ? AND CREATE_TIME < ? ORDER BY VISIT_DATE`, userNum, monthStart, nextMonth)
	if err != nil {
		return nil, err
	}
	visitDates := make([]string, 0, len(rows))
	for _, row := range rows {
		if value := model.StringValue(row, "VISIT_DATE"); value != "" {
			visitDates = append(visitDates, value)
		}
	}
	return monthlyCompanionshipResponse(now, visitDates), nil
}

func normalizeLuluPage(page, pageSize, defaultPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

func clampLuluPage(page, pageSize int, total int64) int {
	totalPages := maxInt(1, int((total+int64(pageSize)-1)/int64(pageSize)))
	return minInt(page, totalPages)
}

func luluPageResponse(items []map[string]any, total int64, page, pageSize int) map[string]any {
	totalPages := maxInt(1, int((total+int64(pageSize)-1)/int64(pageSize)))
	return map[string]any{
		"ITEMS": items, "TOTAL": total, "PAGE": page, "PAGE_SIZE": pageSize, "TOTAL_PAGES": totalPages,
	}
}

func monthlyCompanionshipResponse(now time.Time, visitDates []string) map[string]any {
	uniqueDates := make(map[string]struct{}, len(visitDates))
	monthPrefix := now.Format("2006-01") + "-"
	for _, value := range visitDates {
		if strings.HasPrefix(value, monthPrefix) && value <= now.Format("2006-01-02") {
			uniqueDates[value] = struct{}{}
		}
	}
	visitedDays := len(uniqueDates)
	elapsedDays := now.Day()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
	missedDays := maxInt(0, elapsedDays-visitedDays)
	cleanDates := make([]string, 0, visitedDays)
	for value := range uniqueDates {
		cleanDates = append(cleanDates, value)
	}
	slices.Sort(cleanDates)
	return map[string]any{
		"MONTH": now.Format("2006-01"), "VISITED_DAYS": visitedDays, "MISSED_DAYS": missedDays,
		"ELAPSED_DAYS": elapsedDays, "DAYS_IN_MONTH": daysInMonth, "VISITED_DATES": cleanDates,
	}
}

func (s *Service) GetPetFunState(ctx context.Context, userNum int64) (map[string]any, error) {
	state, err := s.getOrCreateFunState(ctx, userNum)
	if err != nil {
		return nil, err
	}
	refreshFunState(state)
	if err := s.updateFunState(ctx, state); err != nil {
		return nil, err
	}
	return funStateResponse(state), nil
}

func (s *Service) AdvancePetMission(ctx context.Context, userNum int64, missionType string, amount int) (map[string]any, error) {
	state, err := s.getOrCreateFunState(ctx, userNum)
	if err != nil {
		return nil, err
	}
	refreshFunState(state)
	if model.StringValue(state, "MISSION_TYPE") == missionType {
		target := model.IntValue(state, "MISSION_TARGET")
		state["MISSION_CURRENT"] = minInt(target, model.IntValue(state, "MISSION_CURRENT")+maxInt(1, amount))
	}
	if err := s.updateFunState(ctx, state); err != nil {
		return nil, err
	}
	return funStateResponse(state), nil
}

func (s *Service) ChangePetClothes(ctx context.Context, userNum int64) (map[string]any, error) {
	state, err := s.getOrCreateFunState(ctx, userNum)
	if err != nil {
		return nil, err
	}
	refreshFunState(state)
	state["CLOTHES_INDEX"] = model.IntValue(state, "CLOTHES_INDEX") + 1
	if err := s.updateFunState(ctx, state); err != nil {
		return nil, err
	}
	return funStateResponse(state), nil
}

func (s *Service) createDefaultPet(ctx context.Context, userNum int64) (map[string]any, error) {
	now := time.Now()
	pet := map[string]any{
		"USER_NUM": userNum, "NAME": defaultPetName, "HUNGER": 50, "ENERGY": 60, "MOOD": 80,
		"LEVEL": 1, "EXP": 0, "CURRENT_STATE": "IDLE", "LAST_UPDATE_TIME": now,
	}
	_, err := s.Repo.Exec(ctx, `INSERT INTO z_pet_status
(USER_NUM, NAME, HUNGER, ENERGY, MOOD, LEVEL, EXP, CURRENT_STATE, LAST_UPDATE_TIME)
VALUES (?, ?, 50, 60, 80, 1, 0, 'IDLE', ?)`, userNum, defaultPetName, now)
	return pet, err
}

func (s *Service) refreshPet(ctx context.Context, pet map[string]any) error {
	lastUpdate := parseTime(model.Lookup(pet, "LAST_UPDATE_TIME"))
	minutes := int(time.Since(lastUpdate).Minutes())
	if minutes <= 0 {
		return nil
	}
	hunger := model.IntValue(pet, "HUNGER")
	energy := model.IntValue(pet, "ENERGY")
	mood := model.IntValue(pet, "MOOD")
	if model.StringValue(pet, "CURRENT_STATE") == "SLEEPING" {
		energy = minInt(100, energy+minutes*4)
		hunger = maxInt(0, hunger-minutes)
		if energy >= 100 {
			pet["CURRENT_STATE"] = "IDLE"
		}
	} else {
		hunger = maxInt(0, hunger-minutes)
		energy = maxInt(0, energy-minutes)
		if hunger <= 20 || energy <= 20 {
			mood = maxInt(0, mood-minutes)
		}
	}
	pet["HUNGER"], pet["ENERGY"], pet["MOOD"] = hunger, energy, mood
	pet["LAST_UPDATE_TIME"] = lastUpdate.Add(time.Duration(minutes) * time.Minute)
	return s.updatePet(ctx, pet)
}

func (s *Service) updatePet(ctx context.Context, pet map[string]any) error {
	_, err := s.Repo.Exec(ctx, `UPDATE z_pet_status SET HUNGER = ?, ENERGY = ?, MOOD = ?, LEVEL = ?, EXP = ?, CURRENT_STATE = ?, LAST_UPDATE_TIME = ? WHERE USER_NUM = ?`,
		model.IntValue(pet, "HUNGER"), model.IntValue(pet, "ENERGY"), model.IntValue(pet, "MOOD"),
		maxInt(1, model.IntValue(pet, "LEVEL")), model.IntValue(pet, "EXP"), model.StringValue(pet, "CURRENT_STATE"),
		parseTime(model.Lookup(pet, "LAST_UPDATE_TIME")), model.Int64Value(pet, "USER_NUM"))
	return err
}

func addPetExperience(pet map[string]any, amount int) {
	level := maxInt(1, model.IntValue(pet, "LEVEL"))
	experience := model.IntValue(pet, "EXP") + amount
	for experience >= level*50+100 {
		experience -= level*50 + 100
		level++
	}
	pet["LEVEL"], pet["EXP"] = level, experience
}

func (s *Service) insertPetLog(ctx context.Context, userNum int64, actionType, actionName, ip, userAgent, remark string) {
	if strings.TrimSpace(ip) == "" {
		ip = "unknown"
	}
	_, _ = s.Repo.Exec(ctx, `INSERT INTO z_lulu_log
(USER_NUM, ACTION_TYPE, ACTION_NAME, IP_ADDRESS, BROWSER, DEVICE_MODEL, USER_AGENT, REMARK)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, userNum, actionType, actionName, ip, parseBrowser(userAgent), parseDevice(userAgent), truncateRunes(userAgent, 500), remark)
}

func (s *Service) getOrCreateFunState(ctx context.Context, userNum int64) (map[string]any, error) {
	rows, err := s.Repo.Query(ctx, "SELECT * FROM z_lulu_fun_state WHERE USER_NUM = ?", userNum)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return rows[0], nil
	}
	state := defaultFunState(userNum)
	_, err = s.Repo.Exec(ctx, `INSERT INTO z_lulu_fun_state
(USER_NUM, STATE_DATE, LAST_VISIT_DATE, STREAK_COUNT, MISSION_TYPE, MISSION_TITLE, MISSION_TARGET, MISSION_CURRENT, MISSION_REWARD, CLOTHES_INDEX, UPDATE_TIME)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, userNum, state["STATE_DATE"], state["LAST_VISIT_DATE"], 1,
		state["MISSION_TYPE"], state["MISSION_TITLE"], state["MISSION_TARGET"], 0, state["MISSION_REWARD"], 0, state["UPDATE_TIME"])
	return state, err
}

func defaultFunState(userNum int64) map[string]any {
	now := time.Now()
	mission := dailyMissions[now.YearDay()%len(dailyMissions)]
	return map[string]any{
		"USER_NUM": userNum, "STATE_DATE": dateString(now), "LAST_VISIT_DATE": dateString(now), "STREAK_COUNT": 1,
		"MISSION_TYPE": mission[0], "MISSION_TITLE": mission[1], "MISSION_TARGET": parseLimit(mission[2], 1),
		"MISSION_CURRENT": 0, "MISSION_REWARD": mission[3], "CLOTHES_INDEX": 0, "UPDATE_TIME": now,
	}
}

func refreshFunState(state map[string]any) {
	now := time.Now()
	today := dateString(now)
	if model.StringValue(state, "STATE_DATE") != today {
		mission := dailyMissions[now.YearDay()%len(dailyMissions)]
		state["STATE_DATE"], state["MISSION_TYPE"], state["MISSION_TITLE"] = today, mission[0], mission[1]
		state["MISSION_TARGET"], state["MISSION_CURRENT"], state["MISSION_REWARD"] = parseLimit(mission[2], 1), 0, mission[3]
	}
	lastVisit := model.StringValue(state, "LAST_VISIT_DATE")
	if lastVisit != today {
		streak := 1
		if lastVisit == dateString(now.AddDate(0, 0, -1)) {
			streak = maxInt(1, model.IntValue(state, "STREAK_COUNT")) + 1
		}
		state["STREAK_COUNT"], state["LAST_VISIT_DATE"] = streak, today
	}
	state["UPDATE_TIME"] = now
}

func (s *Service) updateFunState(ctx context.Context, state map[string]any) error {
	_, err := s.Repo.Exec(ctx, `UPDATE z_lulu_fun_state SET STATE_DATE = ?, LAST_VISIT_DATE = ?, STREAK_COUNT = ?,
MISSION_TYPE = ?, MISSION_TITLE = ?, MISSION_TARGET = ?, MISSION_CURRENT = ?, MISSION_REWARD = ?, CLOTHES_INDEX = ?, UPDATE_TIME = ?
WHERE USER_NUM = ?`, state["STATE_DATE"], state["LAST_VISIT_DATE"], model.IntValue(state, "STREAK_COUNT"),
		state["MISSION_TYPE"], state["MISSION_TITLE"], model.IntValue(state, "MISSION_TARGET"), model.IntValue(state, "MISSION_CURRENT"),
		state["MISSION_REWARD"], model.IntValue(state, "CLOTHES_INDEX"), time.Now(), model.Int64Value(state, "USER_NUM"))
	return err
}

func funStateResponse(state map[string]any) map[string]any {
	return map[string]any{
		"MISSION_TYPE": model.StringValue(state, "MISSION_TYPE"), "MISSION_TITLE": model.StringValue(state, "MISSION_TITLE"),
		"MISSION_TARGET": model.IntValue(state, "MISSION_TARGET"), "MISSION_CURRENT": model.IntValue(state, "MISSION_CURRENT"),
		"MISSION_REWARD": model.StringValue(state, "MISSION_REWARD"), "STREAK_COUNT": maxInt(1, model.IntValue(state, "STREAK_COUNT")),
		"CLOTHES_INDEX": model.IntValue(state, "CLOTHES_INDEX"),
	}
}

func parseBrowser(userAgent string) string {
	for marker, name := range map[string]string{"Edg/": "Microsoft Edge", "OPR/": "Opera", "Opera": "Opera", "Firefox/": "Firefox", "CriOS/": "Chrome", "Chrome/": "Chrome", "Safari/": "Safari"} {
		if strings.Contains(userAgent, marker) {
			return name
		}
	}
	return "未知浏览器"
}

func parseDevice(userAgent string) string {
	for marker, name := range map[string]string{"iPhone": "iPhone", "iPad": "iPad", "Windows": "Windows", "Macintosh": "Mac"} {
		if strings.Contains(userAgent, marker) {
			return name
		}
	}
	if index := strings.Index(userAgent, "Android"); index >= 0 {
		end := strings.Index(userAgent[index:], ")")
		part := userAgent[index:]
		if end >= 0 {
			part = userAgent[index : index+end]
		}
		pieces := strings.Split(part, ";")
		if len(pieces) >= 3 {
			return strings.TrimSpace(pieces[len(pieces)-1])
		}
		return "Android"
	}
	return "未知设备"
}

func parseTime(value any) time.Time {
	switch typed := value.(type) {
	case time.Time:
		return typed
	case string:
		for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05"} {
			if parsed, err := time.ParseInLocation(layout, strings.Split(typed, ".")[0], time.Local); err == nil {
				return parsed
			}
		}
	}
	return time.Now()
}

func dateString(value time.Time) string { return value.Format("2006-01-02") }
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
