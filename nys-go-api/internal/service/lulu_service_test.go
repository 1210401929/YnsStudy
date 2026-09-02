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
	if event := chooseLuluNPCEvent(1, 99, 0); event != nil {
		t.Fatalf("first-time visitor should not receive NPC event: %#v", event)
	}
	if event := chooseLuluNPCEvent(10, 0, 0); event != nil {
		t.Fatalf("cooldown should suppress NPC event: %#v", event)
	}
	if event := chooseLuluNPCEvent(5, 1, 499); event == nil || event.EventType != "OUTING" {
		t.Fatalf("outing event not selected: %#v", event)
	}
	if event := chooseLuluNPCEvent(2, 1, 500); event == nil || event.EventType != "LETTER" {
		t.Fatalf("letter event not selected: %#v", event)
	}
	if event := chooseLuluNPCEvent(20, 1, 800); event != nil {
		t.Fatalf("ordinary roll should produce no event: %#v", event)
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
