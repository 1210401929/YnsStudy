package service

import (
	"context"
	"errors"
	"slices"

	"nys-go-api/internal/model"
)

var ErrLuluOutfitUnavailable = errors.New("该服饰不存在或角色尚未解锁")

// Frozen order from the old UI: CLOTHES_INDEX was an unbounded click count modulo 13.
var legacyLuluOutfitKeys = [...]string{
	"default", "hoodie", "overalls", "scarf", "crossbody", "pajamas", "greenpants",
	"greenpants-puffer", "watermelon", "shark-slippers", "dinosaur-box", "cockroach", "tie",
}

func validLuluOutfit(key string) bool {
	return slices.Contains(legacyLuluOutfitKeys[:], key) || key == "winter-puffer" || key == "winter-knit"
}

func validLumeiOutfit(key string) bool {
	return key == "default" || key == "winter-puffer" || key == "winter-knit"
}

type LuluOutfitState struct {
	LuluKey       string `json:"LULU_OUTFIT_KEY"`
	LumeiKey      string `json:"LUMEI_OUTFIT_KEY"`
	LumeiUnlocked bool   `json:"LUMEI_UNLOCKED"`
	Revision      int64  `json:"REVISION"`
}

func luluOutfitState(row map[string]any) LuluOutfitState {
	luluKey := model.StringValue(row, "LULU_OUTFIT_KEY")
	if !validLuluOutfit(luluKey) {
		legacyIndex := max(0, model.Int64Value(row, "CLOTHES_INDEX"))
		luluKey = legacyLuluOutfitKeys[legacyIndex%int64(len(legacyLuluOutfitKeys))]
	}
	lumeiKey := model.StringValue(row, "LUMEI_OUTFIT_KEY")
	if !validLumeiOutfit(lumeiKey) {
		lumeiKey = "default"
	}
	return LuluOutfitState{
		LuluKey: luluKey, LumeiKey: lumeiKey,
		LumeiUnlocked: model.IntValue(row, "PET_LEVEL") >= lumeiResidentLevel,
		Revision:      model.Int64Value(row, "REVISION"),
	}
}

func (s *Service) GetLuluOutfits(ctx context.Context) (LuluOutfitState, error) {
	rows, err := s.Repo.Query(ctx, `SELECT outfit.LULU_OUTFIT_KEY, outfit.LUMEI_OUTFIT_KEY, outfit.REVISION,
fun.CLOTHES_INDEX, COALESCE(pet.LEVEL, 1) AS PET_LEVEL
FROM (SELECT ? AS USER_NUM) AS shared_pet
LEFT JOIN z_pet_status AS pet ON pet.USER_NUM = shared_pet.USER_NUM
LEFT JOIN z_lulu_fun_state AS fun ON fun.USER_NUM = shared_pet.USER_NUM
LEFT JOIN z_lulu_outfit_state AS outfit ON outfit.USER_NUM = shared_pet.USER_NUM`, publicLuluUserNum)
	if err != nil {
		return LuluOutfitState{}, err
	}
	return luluOutfitState(rows[0]), nil // The derived row exists even before the pet is created.
}

func (s *Service) ChangeLuluOutfit(ctx context.Context, character, key string) (LuluOutfitState, error) {
	state, err := s.GetLuluOutfits(ctx)
	if err != nil {
		return LuluOutfitState{}, err
	}
	var column string
	switch character {
	case "lulu":
		if !validLuluOutfit(key) {
			return LuluOutfitState{}, ErrLuluOutfitUnavailable
		}
		column, state.LuluKey = "LULU_OUTFIT_KEY", key
	case "lumei":
		if !state.LumeiUnlocked || !validLumeiOutfit(key) {
			return LuluOutfitState{}, ErrLuluOutfitUnavailable
		}
		column, state.LumeiKey = "LUMEI_OUTFIT_KEY", key
	default:
		return LuluOutfitState{}, ErrLuluOutfitUnavailable
	}

	// The insert preserves the other character's current outfit. Once a row exists,
	// update only the chosen character so simultaneous changes cannot overwrite each other.
	_, err = s.Repo.Exec(ctx, `INSERT INTO z_lulu_outfit_state (USER_NUM, LULU_OUTFIT_KEY, LUMEI_OUTFIT_KEY, REVISION)
VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE `+column+` = ?, REVISION = REVISION + 1`,
		publicLuluUserNum, state.LuluKey, state.LumeiKey, key)
	if err != nil {
		return LuluOutfitState{}, err
	}
	return s.GetLuluOutfits(ctx)
}
