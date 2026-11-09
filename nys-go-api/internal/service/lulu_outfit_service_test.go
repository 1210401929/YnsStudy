package service

import "testing"

func TestLuluOutfitStatePreservesLegacySelection(t *testing.T) {
	// 25 old clicks meant index 12 in the 13-outfit UI. Modulo the expanded
	// catalog would instead select another outfit.
	legacy := luluOutfitState(map[string]any{"CLOTHES_INDEX": 25, "PET_LEVEL": 54})
	if legacy.LuluKey != "tie" || legacy.LumeiKey != "default" || legacy.LumeiUnlocked {
		t.Fatalf("unexpected legacy outfit state: %+v", legacy)
	}
	stored := luluOutfitState(map[string]any{
		"CLOTHES_INDEX": 25, "PET_LEVEL": 55, "REVISION": 4,
		"LULU_OUTFIT_KEY": "winter-puffer", "LUMEI_OUTFIT_KEY": "winter-knit",
	})
	want := LuluOutfitState{LuluKey: "winter-puffer", LumeiKey: "winter-knit", LumeiUnlocked: true, Revision: 4}
	if stored != want {
		t.Fatalf("got %+v; want %+v", stored, want)
	}
}
