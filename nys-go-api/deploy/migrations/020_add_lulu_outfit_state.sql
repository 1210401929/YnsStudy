-- Stable selections for the shared Lulu and Lumei wardrobe.
-- Until the first selection is saved, the API reads Lulu's existing CLOTHES_INDEX.
CREATE TABLE IF NOT EXISTS z_lulu_outfit_state (
    USER_NUM         BIGINT      NOT NULL,
    LULU_OUTFIT_KEY   VARCHAR(32) NOT NULL,
    LUMEI_OUTFIT_KEY  VARCHAR(32) NOT NULL,
    REVISION         BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (USER_NUM)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
