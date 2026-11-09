-- One shared wedding per Lulu. The primary key makes the ceremony log idempotent.
CREATE TABLE IF NOT EXISTS z_lulu_wedding (
    USER_NUM   BIGINT   NOT NULL,
    MARRIED_AT DATETIME NOT NULL,
    PRIMARY KEY (USER_NUM)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS z_lulu_wedding_blessing (
    ID           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    USER_NUM     BIGINT          NOT NULL,
    VISITOR_HASH CHAR(64)        NOT NULL,
    BLESSING_DAY DATE            NOT NULL,
    CONTENT      VARCHAR(60)     NOT NULL,
    CREATED_AT   DATETIME        NOT NULL,
    PRIMARY KEY (ID),
    UNIQUE KEY uniq_lulu_wedding_daily_blessing (USER_NUM, VISITOR_HASH, BLESSING_DAY),
    KEY idx_lulu_wedding_latest_blessing (USER_NUM, ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
