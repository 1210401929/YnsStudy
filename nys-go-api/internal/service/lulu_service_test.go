package service

import (
	"testing"
	"time"
)

func TestSelectSmartCarePlan(t *testing.T) {
	tests := []struct {
		name string
		pet  map[string]any
		want string
	}{
		{"keeps sleeping pet asleep", map[string]any{"CURRENT_STATE": "SLEEPING", "HUNGER": 10, "ENERGY": 20, "MOOD": 20}, "RESTING"},
		{"feeds hungry pet first", map[string]any{"CURRENT_STATE": "IDLE", "HUNGER": 40, "ENERGY": 20, "MOOD": 20}, "FEED"},
		{"rests tired pet", map[string]any{"CURRENT_STATE": "IDLE", "HUNGER": 80, "ENERGY": 30, "MOOD": 20}, "REST"},
		{"comforts unhappy pet", map[string]any{"CURRENT_STATE": "IDLE", "HUNGER": 80, "ENERGY": 80, "MOOD": 60}, "COMFORT"},
		{"walks with healthy pet", map[string]any{"CURRENT_STATE": "IDLE", "HUNGER": 80, "ENERGY": 80, "MOOD": 80}, "STROLL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := selectSmartCarePlan(test.pet).Action; got != test.want {
				t.Fatalf("selectSmartCarePlan() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestClampPetStat(t *testing.T) {
	for value, want := range map[int]int{-12: 0, 0: 0, 42: 42, 100: 100, 140: 100} {
		if got := clampPetStat(value); got != want {
			t.Fatalf("clampPetStat(%d) = %d, want %d", value, got, want)
		}
	}
}

func TestLuluPageResponse(t *testing.T) {
	response := luluPageResponse([]map[string]any{{"ID": 1}}, 21, 2, 10)
	if response["TOTAL_PAGES"] != 3 || response["PAGE"] != 2 || response["TOTAL"] != int64(21) {
		t.Fatalf("unexpected page response: %#v", response)
	}
}

func TestMonthlyCompanionshipResponse(t *testing.T) {
	now := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.Local)
	response := monthlyCompanionshipResponse(now, []string{"2026-08-01", "2026-08-01", "2026-08-12", "2026-09-01"})
	if response["VISITED_DAYS"] != 2 || response["MISSED_DAYS"] != 24 || response["ELAPSED_DAYS"] != 26 || response["DAYS_IN_MONTH"] != 31 {
		t.Fatalf("unexpected companionship response: %#v", response)
	}
}

func TestChooseLuluNPCEvent(t *testing.T) {
	if event := chooseLuluNPCEvent(1, 99, 10, 0); event != nil {
		t.Fatalf("first-time visitor should not receive NPC event: %#v", event)
	}
	if event := chooseLuluNPCEvent(10, 0, 10, 0); event != nil {
		t.Fatalf("cooldown should suppress NPC event: %#v", event)
	}
	if event := chooseLuluNPCEvent(5, 1, 10, 119); event == nil || event.EventType != "OUTING" {
		t.Fatalf("outing event not selected: %#v", event)
	}
	if event := chooseLuluNPCEvent(5, 1, 10, 120); event == nil || event.EventType != "LETTER" {
		t.Fatalf("letter event not selected: %#v", event)
	}
	if event := chooseLuluNPCEvent(2, 1, 10, 100); event == nil || event.EventType != "LETTER" {
		t.Fatalf("newer returning visitor should receive a letter instead of an outing: %#v", event)
	}
	if event := chooseLuluNPCEvent(20, 1, 10, 930); event != nil {
		t.Fatalf("ordinary roll should produce no event: %#v", event)
	}
	if event := chooseLuluNPCEvent(20, 1, 45, 950); event == nil || event.EventType != "LETTER" {
		t.Fatalf("story-stage visitor should receive the higher-probability themed letter: %#v", event)
	}
	if event := chooseLuluNPCEvent(20, 1, 55, 200); event != nil {
		t.Fatalf("resident Lumei should stop sending remote letters: %#v", event)
	}
}

func TestLumeiStoryLettersFollowLevel(t *testing.T) {
	if got := lumeiLetterEventForLevel(44, 7).Title; got != "噜妹寄来一封信" {
		t.Fatalf("unexpected ordinary letter title: %q", got)
	}
	if got := lumeiLetterEventForLevel(45, 7).Title; got != "噜妹写下了想留下来的心愿" {
		t.Fatalf("unexpected invitation letter title: %q", got)
	}
	if got := lumeiLetterEventForLevel(50, 7).Title; got != "噜妹寄来一张搬家清单" {
		t.Fatalf("unexpected preparing letter title: %q", got)
	}
	if got := lumeiLetterEventForLevel(54, 7).Title; got != "噜妹寄来入住倒计时" {
		t.Fatalf("unexpected countdown letter title: %q", got)
	}
}

func TestChooseLumeiSoloOuting(t *testing.T) {
	if event := chooseLumeiSoloOuting(99, 1, 0); event != nil {
		t.Fatalf("Lumei should stay home during the first two resident days: %#v", event)
	}
	if event := chooseLumeiSoloOuting(1, 10, 0); event != nil {
		t.Fatalf("solo outing cooldown should be respected: %#v", event)
	}
	if event := chooseLumeiSoloOuting(10, 10, lumeiSoloOutingRollLimit); event != nil {
		t.Fatalf("high roll should not trigger solo outing: %#v", event)
	}
	if event := chooseLumeiSoloOuting(10, 10, lumeiSoloOutingRollLimit-1); event == nil || event.EventType != "LUMEI_OUTING" {
		t.Fatalf("solo outing should be selected: %#v", event)
	}
}

func TestLumeiWeeklyTaskSpecs(t *testing.T) {
	now := time.Date(2026, time.September, 14, 12, 0, 0, 0, time.Local)
	first := lumeiWeeklyTaskSpecsForDate(now)
	second := lumeiWeeklyTaskSpecsForDate(now.AddDate(0, 0, 2))
	if len(first) < 1 || len(first) > 2 || len(first) != len(second) {
		t.Fatalf("weekly task count should stay between one and two: %#v", first)
	}
	for index := range first {
		if first[index].TaskType != second[index].TaskType || first[index].Target <= 0 || first[index].Condition == "" {
			t.Fatalf("weekly tasks should be valid and stable: %#v / %#v", first, second)
		}
	}
}

func TestLumeiResidencyResponse(t *testing.T) {
	tests := []struct {
		level        int
		wantStatus   string
		wantNext     int
		wantResident bool
	}{
		{1, "NPC", 45, false},
		{45, "INVITATION", 50, false},
		{50, "PREPARING", 54, false},
		{54, "COUNTDOWN", 55, false},
		{55, "RESIDENT", 0, true},
		{80, "RESIDENT", 0, true},
	}
	for _, test := range tests {
		response := lumeiResidencyResponse(test.level)
		if response["STATUS"] != test.wantStatus || response["NEXT_MILESTONE_LEVEL"] != test.wantNext || response["RESIDENT"] != test.wantResident {
			t.Fatalf("lumeiResidencyResponse(%d) = %#v", test.level, response)
		}
		progress, ok := response["PROGRESS"].(int)
		if !ok || progress < 0 || progress > 100 {
			t.Fatalf("invalid progress at level %d: %#v", test.level, response["PROGRESS"])
		}
	}
}

func TestLumeiLetterContents(t *testing.T) {
	if got, want := len(lumeiLetterContents), 19; got != want {
		t.Fatalf("len(lumeiLetterContents) = %d, want %d", got, want)
	}
	seen := make(map[string]struct{}, len(lumeiLetterContents))
	for _, content := range lumeiLetterContents {
		if content == "" {
			t.Fatal("letter content should not be empty")
		}
		if _, exists := seen[content]; exists {
			t.Fatalf("duplicate letter content: %q", content)
		}
		seen[content] = struct{}{}
	}
}

func TestStableLuluMemorySignature(t *testing.T) {
	first := stableLuluMemorySignature("192.0.2.10")
	if first == "" || first != stableLuluMemorySignature("192.0.2.10") {
		t.Fatalf("memory signature should be non-empty and stable: %q", first)
	}
}

func TestStableLuluEventRoll(t *testing.T) {
	first := stableLuluEventRoll("192.0.2.10", "2026-08-28")
	if first < 0 || first >= luluNPCEventRollMax {
		t.Fatalf("event roll out of range: %d", first)
	}
	if first != stableLuluEventRoll("192.0.2.10", "2026-08-28") {
		t.Fatalf("event roll should stay stable for the same IP and day")
	}
}

func TestLuluCommunityGoalForDate(t *testing.T) {
	goal := luluCommunityGoalForDate(time.Date(2026, time.August, 28, 12, 0, 0, 0, time.Local))
	if goal.GoalType == "" || goal.Target <= 0 || goal.Condition == "" {
		t.Fatalf("unexpected community goal: %#v", goal)
	}
}
