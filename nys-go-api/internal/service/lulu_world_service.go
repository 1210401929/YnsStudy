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
	luluNPCOutingRollLimit   = 120
	luluNPCLetterRollLimit   = 930
	luluNPCEventCooldownDays = 1
	lumeiResidentLevel       = 55
	lumeiSoloOutingRollLimit = 180
	lumeiSoloOutingCooldown  = 2
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

type lumeiWeeklyTaskSpec struct {
	TaskType    string
	Title       string
	Description string
	Target      int
	Condition   string
	Reward      string
}

var lumeiLetterContents = []string{
	"我在路边看见一朵像小太阳的花，第一时间就想送给你。今天也要开心呀！",
	"今天的云像一大块软软的棉花糖。替我告诉噜噜，下次要一起去看。",
	"我偷偷准备了新的游戏，下次见面要和噜噜一决胜负。先帮我保密哦！",
	"路过小面包店时闻到了橘子香，我猜噜噜一定会喜欢，所以写信来提醒它记得吃饭。",
	"我把今天最圆的一片云画在了信纸背面。噜噜看到时，会不会说它像一颗大橘子？",
	"院子里来了一只不怕人的小麻雀，我和它聊了好久。下次让噜噜也来认识它吧。",
	"刚才风把我的蝴蝶结吹歪了，我整理好时突然想到：噜噜今天有没有乖乖梳头呀？",
	"我存下了一颗特别甜的糖，没有偷偷吃掉哦。等见到噜噜，我们一人一半。",
	"今天学会了一个很厉害的纸飞机折法，下次比赛看谁的飞机能飞过小树！",
	"傍晚的天空变成了橘子汽水的颜色，真想和噜噜肩并肩坐着看一会儿。",
	"听说噜噜最近被照顾得很好，我也跟着开心起来。谢谢你常常陪着它呀！",
	"我在旧盒子里找到一枚亮晶晶的纽扣，决定把它留给噜噜当幸运宝物。",
	"今天走路时踩到一片会嘎吱响的叶子，我来回踩了三次，想留一次给噜噜。",
	"我做了一个梦，梦里噜噜坐着橘子小船，我们一起划过了软绵绵的云海。",
	"路边的小猫冲我眨了一下眼睛。我猜它也认识噜噜，下次一定要问清楚。",
	"我新学会了一首短短的歌，虽然有一句总唱跑调，但噜噜应该不会笑我吧？",
	"今天没有什么大事，只是忽然很想念噜噜，所以就写了这封小小的信。",
	"我发现快乐也可以攒起来：一颗糖、一阵风，还有等着下次见噜噜的期待。",
	"请替我检查一下噜噜有没有按时休息。要是它还在打哈欠，就帮我劝它睡个好觉。",
}

var lumeiInvitationLetterContents = []string{
	"最近每次要回去时，我都会忍不住回头看噜噜的小屋。如果能多住几天就好啦。",
	"我给噜噜画了一张小房间的图，还悄悄画了两把椅子。你猜另一把是给谁的？",
	"噜噜说这里总会有人来陪它，听起来真温暖。我也可以把这里当成第二个家吗？",
	"我收好了一只小杯子，下次来时想把它留在噜噜身边，这样就不用每次带来带去啦。",
	"昨天梦见我和噜噜在同一个房间里醒来，然后一起等你。醒来以后还有一点舍不得。",
	"如果我以后常常住下，我会负责提醒噜噜按时吃饭，也会记得每天和你打招呼。",
}

var lumeiPreparingLetterContents = []string{
	"噜噜说已经帮我留好了一个小角落，我正在挑一只最软的枕头。",
	"今天装箱时翻出了好多旧信，原来我们已经有这么多回忆了。我会把它们全部带来。",
	"我想在新房间摆两只杯子，一只给噜噜，一只给经常来看我们的你。",
	"搬家清单：花边小被子、橘子杯、和噜噜玩的球，还有一大袋期待。",
	"我练习了好几次「我回来啦」，结果每次都笑场。到时候你可不许笑我哦。",
	"噜噜已经把房间量了三遍，生怕我的箱子放不下。其实我最想带来的只有回忆。",
}

var lumeiCountdownLetterContents = []string{
	"最后一只箱子已经合上啦！下次见面，我可能就不是来做客的了。",
	"我把钥匙挂绳编好了，是粉色的。噜噜说家里的那把新钥匙已经在等我。",
	"今晚可能会兴奋得睡不着，因为再往前一点点，我们就不用再说「下次见」啦。",
	"请帮我告诉噜噜：我马上就到，让它不要又在门口紧张地走来走去。",
	"这可能是我从远方寄来的最后几封信之一。以后有话，我想当面说给你听。",
}

var lumeiSoloOutingContents = []string{
	"噜妹留下小便签：「我去花店挑一盆小花，噜噜会在家陪你，很快回来。」",
	"噜妹背着小包去买橘子了，桌上压着一张写给你和噜噜的便签。",
	"噜妹去取新窗帘啦。她特意叮嘱：「我不在时，也要好好陪噜噜哦。」",
	"噜妹拎着小篮子去市集了，回来时会带一份神秘小点心。",
	"噜妹去给老朋友送一封信，这次噜噜没有跟去，正在家里等你。",
}

var lumeiWeeklyTaskPool = []lumeiWeeklyTaskSpec{
	{TaskType: "FEED", Title: "双人点心准备周", Description: "和噜妹一起给噜噜准备 6 顿好吃的。", Target: 6, Condition: "ACTION_TYPE = 'FEED'", Reward: "解锁一颗橘子点心星"},
	{TaskType: "PLAY", Title: "噜噜噜妹游戏周", Description: "陪他们完成 5 次玩耍或小游戏。", Target: 5, Condition: "ACTION_TYPE = 'PLAY'", Reward: "解锁双人闪亮心情"},
	{TaskType: "GOOD_NIGHT", Title: "一起说晚安", Description: "本周陪噜噜入睡或醒来 4 次。", Target: 4, Condition: "ACTION_TYPE IN ('SLEEP', 'WAKE')", Reward: "收藏一枚晚安小月亮"},
	{TaskType: "CARE", Title: "双人照顾小队", Description: "让智能照顾帮噜噜和噜妹完成 4 件小事。", Target: 4, Condition: "ACTION_TYPE = 'AUTO_CARE'", Reward: "获得本周照顾小能手印记"},
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
	pet, err := s.GetPetStatus(ctx, userNum)
	if err != nil {
		return nil, err
	}
	petLevel := maxInt(1, model.IntValue(pet, "LEVEL"))
	memory, visitDays, err := s.getLuluMemory(ctx, userNum, ip, now)
	if err != nil {
		return nil, err
	}
	residency, err := s.getLumeiResidency(ctx, userNum, petLevel, now)
	if err != nil {
		return nil, err
	}
	residentSince := time.Time{}
	if value, ok := residency["RESIDENT_SINCE"]; ok {
		residentSince = parseTime(value)
	}
	event, history, err := s.resolveLuluNPCEvent(ctx, userNum, ip, userAgent, visitDays, petLevel, residentSince, now)
	if err != nil {
		return nil, err
	}
	weeklyTasks, err := s.getLumeiWeeklyTasks(ctx, userNum, ip, petLevel, now)
	if err != nil {
		return nil, err
	}
	goal, err := s.getLuluCommunityGoal(ctx, userNum, now)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"NPC_EVENT":       event,
		"NPC_HISTORY":     history,
		"MEMORY":          memory,
		"COMMUNITY_GOAL":  goal,
		"LUMEI_RESIDENCY": residency,
		"LUMEI_WEEKLY":    weeklyTasks,
	}, nil
}

func (s *Service) getLumeiResidency(ctx context.Context, userNum int64, petLevel int, now time.Time) (map[string]any, error) {
	residency := lumeiResidencyResponse(petLevel)
	if petLevel < lumeiResidentLevel {
		return residency, nil
	}

	// The public Lulu has one shared move-in moment. Reusing the existing log table
	// keeps the milestone visible in the ordinary care log without adding a schema.
	_, err := s.Repo.Exec(ctx, `INSERT INTO z_lulu_log
(USER_NUM, ACTION_TYPE, ACTION_NAME, IP_ADDRESS, BROWSER, DEVICE_MODEL, USER_AGENT, REMARK)
SELECT ?, 'NPC_RESIDENT', '噜妹正式入住', 'system', '噜噜世界', '全站事件', '', '噜噜到达 55 级，噜妹带着行李正式搬来常住啦！'
WHERE NOT EXISTS (
	SELECT 1 FROM z_lulu_log WHERE USER_NUM = ? AND ACTION_TYPE = 'NPC_RESIDENT'
)`, userNum, userNum)
	if err != nil {
		return nil, err
	}

	rows, err := s.Repo.Query(ctx, `SELECT MIN(CREATE_TIME) AS RESIDENT_SINCE
FROM z_lulu_log WHERE USER_NUM = ? AND ACTION_TYPE = 'NPC_RESIDENT'`, userNum)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 && model.Lookup(rows[0], "RESIDENT_SINCE") != nil {
		residentSince := parseTime(model.Lookup(rows[0], "RESIDENT_SINCE"))
		if !residentSince.IsZero() {
			residency["RESIDENT_SINCE"] = residentSince.Format("2006-01-02")
		}
	}
	if _, exists := residency["RESIDENT_SINCE"]; !exists {
		residency["RESIDENT_SINCE"] = now.Format("2006-01-02")
	}
	return residency, nil
}

func lumeiResidencyResponse(petLevel int) map[string]any {
	petLevel = maxInt(1, petLevel)
	remainingLevels := maxInt(0, lumeiResidentLevel-petLevel)
	progress := minInt(100, petLevel*100/lumeiResidentLevel)
	status := "NPC"
	title := "噜妹还住在远方"
	message := "她会偶尔寄信，也会悄悄来找噜噜玩。"
	nextLevel := 45
	nextText := "45 级时，噜妹会说出想留下来的心愿。"

	switch {
	case petLevel >= lumeiResidentLevel:
		status = "RESIDENT"
		title = "噜妹的小家完成啦"
		message = "椅子、杯子、床铺和小屋都准备好了，噜妹已经正式住在噜噜身边。"
		nextLevel = 0
		nextText = "入住故事已解锁，点击噜妹可以和她说说话。"
	case petLevel >= 54:
		status = "COUNTDOWN"
		title = "噜妹的行李到了"
		message = "最后一只行李箱已经放在床边，再升 1 级就能正式入住。"
		nextLevel = 55
		nextText = "55 级解锁噜妹常驻。"
	case petLevel >= 53:
		status = "PREPARING"
		title = "小房子的屋顶搭好了"
		message = "噜噜给床搭上了小屋形的屋顶和花边帘子，房间越来越像家。"
		nextLevel = 54
		nextText = "54 级，噜妹的最后一只行李箱会送到。"
	case petLevel >= 52:
		status = "PREPARING"
		title = "软软的床铺好了"
		message = "奶油色枕头和粉色格纹被子已经铺好，噜妹随时都能睡个好觉。"
		nextLevel = 53
		nextText = "53 级，为床搭好小屋形的屋顶。"
	case petLevel >= 51:
		status = "PREPARING"
		title = "两只杯子摆好了"
		message = "小圆桌上放着一只橘色杯子和一只粉色杯子，以后可以一起喝热饮。"
		nextLevel = 52
		nextText = "52 级，为噜妹铺好床和被子。"
	case petLevel >= 50:
		status = "PREPARING"
		title = "第一把椅子到位"
		message = "噜噜先搬来一把粉色软垫椅，噜妹终于有了自己的位置。"
		nextLevel = 51
		nextText = "51 级，添上小圆桌和两只杯子。"
	case petLevel >= 45:
		status = "INVITATION"
		title = "噜妹想把这里当成第二个家"
		message = "她在信里问：「以后我可以常常住在这里吗？」"
		nextLevel = 50
		nextText = "50 级开始准备噜妹的小房间。"
	}

	homeStage := lumeiHomeStageResponse(petLevel)

	return map[string]any{
		"STATUS":               status,
		"RESIDENT":             petLevel >= lumeiResidentLevel,
		"CURRENT_LEVEL":        petLevel,
		"TARGET_LEVEL":         lumeiResidentLevel,
		"REMAINING_LEVELS":      remainingLevels,
		"PROGRESS":             progress,
		"TITLE":                title,
		"MESSAGE":              message,
		"NEXT_MILESTONE_LEVEL": nextLevel,
		"NEXT_MILESTONE_TEXT":  nextText,
		"HOME_STAGE":           homeStage,
		"CHAPTERS": []map[string]any{
			{"LEVEL": 45, "TITLE": "想留下来的信", "DESCRIPTION": "噜妹第一次说出想把这里当成第二个家。", "UNLOCKED": petLevel >= 45, "CURRENT": petLevel >= 45 && petLevel < 50},
			{"LEVEL": 50, "TITLE": "第一把椅子", "DESCRIPTION": "粉色软垫椅先搬进了空房间。", "UNLOCKED": petLevel >= 50, "CURRENT": petLevel == 50},
			{"LEVEL": 51, "TITLE": "小桌与杯子", "DESCRIPTION": "桌上摆好一橘一粉两只杯子。", "UNLOCKED": petLevel >= 51, "CURRENT": petLevel == 51},
			{"LEVEL": 52, "TITLE": "床和软被", "DESCRIPTION": "枕头与粉色格纹被子已经铺好。", "UNLOCKED": petLevel >= 52, "CURRENT": petLevel == 52},
			{"LEVEL": 53, "TITLE": "小屋成形", "DESCRIPTION": "床边搭起小屋形屋顶和花边帘子。", "UNLOCKED": petLevel >= 53, "CURRENT": petLevel == 53},
			{"LEVEL": 54, "TITLE": "最后的行李", "DESCRIPTION": "行李箱和搬家纸箱已经送到床边。", "UNLOCKED": petLevel >= 54, "CURRENT": petLevel == 54},
			{"LEVEL": 55, "TITLE": "完整的小家", "DESCRIPTION": "全部物品归位，噜妹正式成为常驻伙伴。", "UNLOCKED": petLevel >= 55, "CURRENT": petLevel >= 55},
		},
	}
}

func lumeiHomeStageResponse(petLevel int) map[string]any {
	level := 0
	title := "小房间还在计划中"
	description := "到达 50 级后，噜噜会开始一件件准备噜妹的新家。"
	items := []string{}

	switch {
	case petLevel >= 55:
		level = 55
		title = "完整的小家"
		description = "所有家具和生活用品都已归位，噜妹正式入住。"
		items = []string{"粉色软垫椅", "小圆桌与两只杯子", "床与格纹被子", "小屋形床顶", "灯、绿植与拖鞋"}
	case petLevel >= 54:
		level = 54
		title = "行李已经到达"
		description = "家具全部准备好，噜妹的行李箱和纸箱也送到了。"
		items = []string{"粉色软垫椅", "小圆桌与两只杯子", "床与格纹被子", "小屋形床顶", "行李箱与纸箱"}
	case petLevel >= 53:
		level = 53
		title = "小屋已经成形"
		description = "床边搭好了小屋形屋顶和花边帘子。"
		items = []string{"粉色软垫椅", "小圆桌与两只杯子", "床与格纹被子", "小屋形床顶"}
	case petLevel >= 52:
		level = 52
		title = "床和被子准备好了"
		description = "柔软的枕头与粉色格纹被子已经铺好。"
		items = []string{"粉色软垫椅", "小圆桌与两只杯子", "床与格纹被子"}
	case petLevel >= 51:
		level = 51
		title = "桌子和杯子准备好了"
		description = "房间里多了一张小圆桌和两只专属杯子。"
		items = []string{"粉色软垫椅", "小圆桌与两只杯子"}
	case petLevel >= 50:
		level = 50
		title = "第一把椅子准备好了"
		description = "空房间里先放进了一把属于噜妹的粉色软垫椅。"
		items = []string{"粉色软垫椅"}
	}

	return map[string]any{
		"LEVEL":       level,
		"TITLE":       title,
		"DESCRIPTION": description,
		"ITEMS":       items,
		"COMPLETE":    petLevel >= lumeiResidentLevel,
	}
}

func (s *Service) getLumeiWeeklyTasks(ctx context.Context, userNum int64, ip string, petLevel int, now time.Time) (map[string]any, error) {
	weekStart := luluWeekStart(now)
	nextWeek := weekStart.AddDate(0, 0, 7)
	response := map[string]any{
		"UNLOCKED":  petLevel >= lumeiResidentLevel,
		"WEEK_KEY":  weekStart.Format("2006-01-02"),
		"WEEK_START": weekStart.Format("2006-01-02"),
		"WEEK_END":   nextWeek.AddDate(0, 0, -1).Format("2006-01-02"),
		"TASKS":      []map[string]any{},
	}
	if petLevel < lumeiResidentLevel {
		return response, nil
	}

	specs := lumeiWeeklyTaskSpecsForDate(now)
	tasks := make([]map[string]any, 0, len(specs))
	completedCount := 0
	for index, spec := range specs {
		query := `SELECT COUNT(1) AS TOTAL FROM z_lulu_log
WHERE USER_NUM = ? AND IP_ADDRESS = ? AND CREATE_TIME >= ? AND CREATE_TIME < ? AND ` + spec.Condition
		rows, err := s.Repo.Query(ctx, query, userNum, ip, weekStart, nextWeek)
		if err != nil {
			return nil, err
		}
		current := int(firstCount(rows))
		completed := current >= spec.Target
		if completed {
			completedCount++
		}
		tasks = append(tasks, map[string]any{
			"ID":          weekStart.Format("20060102") + "-" + spec.TaskType,
			"TASK_TYPE":   spec.TaskType,
			"TITLE":       spec.Title,
			"DESCRIPTION": spec.Description,
			"CURRENT":     current,
			"TARGET":      spec.Target,
			"REMAINING":   maxInt(0, spec.Target-current),
			"COMPLETED":   completed,
			"REWARD":      spec.Reward,
			"ORDER":       index + 1,
		})
	}
	response["TASKS"] = tasks
	response["COMPLETED_COUNT"] = completedCount
	response["TOTAL_COUNT"] = len(tasks)
	response["ALL_COMPLETED"] = len(tasks) > 0 && completedCount == len(tasks)
	return response, nil
}

func luluWeekStart(now time.Time) time.Time {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekday := int(dayStart.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return dayStart.AddDate(0, 0, -(weekday - 1))
}

func lumeiWeeklyTaskSpecsForDate(now time.Time) []lumeiWeeklyTaskSpec {
	isoYear, isoWeek := now.ISOWeek()
	count := 1 + isoWeek%2
	start := (isoYear + isoWeek*3) % len(lumeiWeeklyTaskPool)
	result := make([]lumeiWeeklyTaskSpec, 0, count)
	for offset := 0; offset < count; offset++ {
		result = append(result, lumeiWeeklyTaskPool[(start+offset*2)%len(lumeiWeeklyTaskPool)])
	}
	return result
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

func (s *Service) resolveLuluNPCEvent(ctx context.Context, userNum int64, ip, userAgent string, visitDays, petLevel int, residentSince, now time.Time) (any, []map[string]any, error) {
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
		roll := stableLuluEventRoll(ip, today)
		var spec *luluNPCEventSpec
		if petLevel >= lumeiResidentLevel {
			daysSinceMoveIn := int(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Sub(residentSince).Hours() / 24)
			spec = chooseLumeiSoloOuting(daysSinceLast, daysSinceMoveIn, roll)
		} else {
			spec = chooseLuluNPCEvent(visitDays, daysSinceLast, petLevel, roll)
		}
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
				} else if spec.EventType == "LUMEI_OUTING" {
					actionType, actionName = "NPC_SOLO_OUTING", "噜妹独自出门"
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
		eventType := model.StringValue(todayRows[0], "EVENT_TYPE")
		if petLevel < lumeiResidentLevel || eventType == "LUMEI_OUTING" {
			currentEvent = luluNPCEventResponse(todayRows[0], now)
		}
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

func chooseLuluNPCEvent(visitDays, daysSinceLast, petLevel, roll int) *luluNPCEventSpec {
	if visitDays < 2 || petLevel >= lumeiResidentLevel || daysSinceLast < luluNPCEventCooldownDays || roll < 0 || roll >= luluNPCEventRollMax {
		return nil
	}
	outingLimit := luluNPCOutingRollLimit
	letterLimit := luluNPCLetterRollLimit
	if petLevel >= 45 {
		outingLimit = 70
		letterLimit = 980
	}
	if visitDays >= 5 && roll < outingLimit {
		return &luluNPCEventSpec{
			EventType: "OUTING",
			Title:     "他们悄悄出门啦",
			Content:   "噜噜和噜妹背着小包去公园找云朵形状了，大约十分钟后就会回来。",
			Duration:  10 * time.Minute,
		}
	}
	if roll < letterLimit {
		return lumeiLetterEventForLevel(petLevel, roll)
	}
	return nil
}

func lumeiLetterEventForLevel(petLevel, roll int) *luluNPCEventSpec {
	title := "噜妹寄来一封信"
	contents := lumeiLetterContents
	duration := 6 * time.Hour
	switch {
	case petLevel >= 54:
		title = "噜妹寄来入住倒计时"
		contents = lumeiCountdownLetterContents
		duration = 12 * time.Hour
	case petLevel >= 50:
		title = "噜妹寄来一张搬家清单"
		contents = lumeiPreparingLetterContents
		duration = 8 * time.Hour
	case petLevel >= 45:
		title = "噜妹写下了想留下来的心愿"
		contents = lumeiInvitationLetterContents
		duration = 8 * time.Hour
	}
	return &luluNPCEventSpec{
		EventType: "LETTER",
		Title:     title,
		Content:   contents[roll%len(contents)],
		Duration:  duration,
	}
}

func chooseLumeiSoloOuting(daysSinceLast, daysSinceMoveIn, roll int) *luluNPCEventSpec {
	if daysSinceMoveIn < 2 || daysSinceLast < lumeiSoloOutingCooldown || roll < 0 || roll >= lumeiSoloOutingRollLimit {
		return nil
	}
	return &luluNPCEventSpec{
		EventType: "LUMEI_OUTING",
		Title:     "噜妹留下一张出门便签",
		Content:   lumeiSoloOutingContents[roll%len(lumeiSoloOutingContents)],
		Duration:  90 * time.Minute,
	}
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
