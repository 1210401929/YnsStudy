package service

import (
	"context"
	"errors"

	"nys-go-api/internal/model"
)

var ErrLuluSceneUnavailable = errors.New("该场景不存在或尚未解锁")

const publicLuluUserNum int64 = 1

var luluScenes = [...]struct {
	key      string
	minLevel int
}{
	{key: "original"},
	{key: "cozy-room"},
	{key: "sunny-garden"},
	{key: "forest-clearing"},
	{key: "seaside-terrace"},
	{key: "starry-camp"},
	{key: "lumei-home", minLevel: lumeiResidentLevel},
	{key: "winter-courtyard"},
	{key: "winter-fireplace"},
	{key: "wedding-room", minLevel: luluWeddingLevel},
}

type LuluSceneState struct {
	SceneKey     string `json:"SCENE_KEY"`
	NextSceneKey string `json:"NEXT_SCENE_KEY"`
	Revision     int64  `json:"REVISION"`
}

type LuluSceneChange struct {
	LuluSceneState
	Applied bool `json:"APPLIED"`
}

func availableLuluSceneIndex(key string, level int) int {
	for i, scene := range luluScenes {
		if scene.key == key && level >= scene.minLevel {
			return i
		}
	}
	return -1
}

func luluSceneState(key string, revision int64, level int) LuluSceneState {
	current := max(0, availableLuluSceneIndex(key, level))
	next := (current + 1) % len(luluScenes)
	for level < luluScenes[next].minLevel {
		next = (next + 1) % len(luluScenes)
	}
	return LuluSceneState{SceneKey: luluScenes[current].key, NextSceneKey: luluScenes[next].key, Revision: revision}
}

func (s *Service) GetLuluScene(ctx context.Context) (LuluSceneState, error) {
	state, _, err := s.getLuluScene(ctx)
	return state, err
}

func (s *Service) getLuluScene(ctx context.Context) (LuluSceneState, int, error) {
	rows, err := s.Repo.Query(ctx, `SELECT scene.SCENE_KEY, scene.REVISION, COALESCE(pet.LEVEL, 1) AS PET_LEVEL
FROM (SELECT ? AS USER_NUM) AS shared_pet
LEFT JOIN z_pet_status AS pet ON pet.USER_NUM = shared_pet.USER_NUM
LEFT JOIN z_lulu_scene_state AS scene ON scene.USER_NUM = shared_pet.USER_NUM`, publicLuluUserNum)
	if err != nil {
		return LuluSceneState{}, 0, err
	}
	row := rows[0] // The derived row exists even before the pet or scene is created.
	level := max(1, model.IntValue(row, "PET_LEVEL"))
	return luluSceneState(model.StringValue(row, "SCENE_KEY"), model.Int64Value(row, "REVISION"), level), level, nil
}

func (s *Service) ChangeLuluScene(ctx context.Context, key string, expectedRevision int64) (LuluSceneChange, error) {
	state, level, err := s.getLuluScene(ctx)
	if err != nil {
		return LuluSceneChange{}, err
	}
	if state.Revision != expectedRevision {
		return LuluSceneChange{LuluSceneState: state}, nil
	}
	if availableLuluSceneIndex(key, level) < 0 {
		return LuluSceneChange{}, ErrLuluSceneUnavailable
	}
	if state.SceneKey == key {
		return LuluSceneChange{LuluSceneState: state, Applied: true}, nil
	}
	if state.Revision == 0 {
		_, err := s.Repo.Exec(ctx, `INSERT INTO z_lulu_scene_state (USER_NUM) VALUES (?)
ON DUPLICATE KEY UPDATE USER_NUM = USER_NUM`, publicLuluUserNum)
		if err != nil {
			return LuluSceneChange{}, err
		}
	}
	// Two visitors may have loaded the same scene. An older selection must not overwrite a newer one.
	affected, err := s.Repo.Exec(ctx, `UPDATE z_lulu_scene_state
SET SCENE_KEY = ?, REVISION = REVISION + 1 WHERE USER_NUM = ? AND REVISION = ?`, key, publicLuluUserNum, expectedRevision)
	if err != nil {
		return LuluSceneChange{}, err
	}
	state, err = s.GetLuluScene(ctx)
	return LuluSceneChange{LuluSceneState: state, Applied: affected == 1}, err
}
