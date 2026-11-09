package service

import "testing"

func TestLuluSceneState(t *testing.T) {
	tests := []struct {
		name, storedKey   string
		level             int
		wantKey, wantNext string
	}{
		{"new site uses the original background", "", 1, "original", "cozy-room"},
		{"snowy courtyard is available before move-in", "starry-camp", 54, "starry-camp", "winter-courtyard"},
		{"shared level 55 unlocks Lumei's home", "starry-camp", 55, "starry-camp", "lumei-home"},
		{"snowy courtyard leads indoors before move-in", "winter-courtyard", 1, "winter-courtyard", "winter-fireplace"},
		{"saved fireplace scene wraps to original", "winter-fireplace", 1, "winter-fireplace", "original"},
		{"wedding room is skipped at Lv.69", "winter-fireplace", 69, "winter-fireplace", "original"},
		{"a saved wedding room is hidden below Lv.70", "wedding-room", 69, "original", "cozy-room"},
		{"wedding room joins the rotation at Lv.70", "winter-fireplace", 70, "winter-fireplace", "wedding-room"},
		{"a saved wedding room stays selected after Lv.70", "wedding-room", 78, "wedding-room", "original"},
		{"unrecognized saved scene falls back", "removed-scene", 55, "original", "cozy-room"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := luluSceneState(test.storedKey, 3, test.level)
			want := LuluSceneState{SceneKey: test.wantKey, NextSceneKey: test.wantNext, Revision: 3}
			if got != want {
				t.Fatalf("got %+v; want %+v", got, want)
			}
		})
	}
}
