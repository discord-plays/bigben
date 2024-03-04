CREATE TABLE bongs
(
    id             BIGINT UNIQUE PRIMARY KEY AUTO_INCREMENT,
    guild_id       BIGINT UNSIGNED NOT NULL,
    user_id        BIGINT UNSIGNED NOT NULL,
    message_id     BIGINT UNSIGNED NOT NULL,
    interaction_id BIGINT UNSIGNED NOT NULL,
    won            BOOLEAN         NOT NULL,
    speed          INTEGER         NOT NULL
);

CREATE TABLE guilds
(
    id                 BIGINT UNSIGNED UNIQUE PRIMARY KEY,
    bong_channel_id    BIGINT UNSIGNED NOT NULL,
    bong_webhook_id    BIGINT UNSIGNED NOT NULL,
    bong_webhook_token TEXT            NOT NULL,
    bong_role_id       BIGINT UNSIGNED NOT NULL,
    bong_emoji         TEXT            NOT NULL
);

CREATE TABLE roles
(
    id         BIGINT UNIQUE PRIMARY KEY AUTO_INCREMENT,
    guild_id   BIGINT UNSIGNED NOT NULL,
    message_id BIGINT UNSIGNED NOT NULL,
    role_id    BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED NOT NULL
);

CREATE TABLE users
(
    id  BIGINT UNSIGNED UNIQUE PRIMARY KEY,
    tag TEXT NOT NULL
);
