-- Shared background for each Lulu, keyed like z_pet_status. No browser data is imported.
CREATE TABLE IF NOT EXISTS z_lulu_scene_state (
    USER_NUM  BIGINT      NOT NULL,
    SCENE_KEY VARCHAR(32) NOT NULL DEFAULT 'original',
    REVISION  BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (USER_NUM)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
