CREATE TABLE bong_log
(
    id             INTEGER UNIQUE PRIMARY KEY AUTO_INCREMENT,
    guild_id       INTEGER UNSIGNED NOT NULL,
    user_id        INTEGER UNSIGNED NOT NULL,
    message_id     INTEGER UNSIGNED NOT NULL,
    interaction_id INTEGER UNSIGNED NOT NULL,
    won            BOOLEAN          NOT NULL,
    speed          INTEGER          NOT NULL
);

CREATE TABLE guild_settings
(
    id                 INTEGER UNIQUE PRIMARY KEY AUTO_INCREMENT,
    guild_id           INTEGER UNSIGNED NOT NULL,
    bong_channel_id    INTEGER UNSIGNED NOT NULL,
    bong_webhook_id    INTEGER UNSIGNED NOT NULL,
    bong_webhook_token TEXT             NOT NULL,
    bong_role_id       INTEGER UNSIGNED NOT NULL,
    bong_emoji         TEXT             NOT NULL
);

CREATE TABLE role_log
(
    id         INTEGER UNIQUE PRIMARY KEY AUTO_INCREMENT,
    guild_id   INTEGER UNSIGNED NOT NULL,
    message_id INTEGER UNSIGNED NOT NULL,
    role_id    INTEGER UNSIGNED NOT NULL,
    user_id    INTEGER UNSIGNED NOT NULL
);

CREATE TABLE user_log
(
    id  INTEGER UNSIGNED UNIQUE PRIMARY KEY,
    tag TEXT NOT NULL
);
