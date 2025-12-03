CREATE TABLE IF NOT EXISTS logins (
                                      id            bigserial PRIMARY KEY,
                                      login         text NOT NULL,
                                      password_hash text NOT NULL,
                                      created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_logins_login ON logins (login);

CREATE TABLE IF NOT EXISTS passwords (
                                         user_id     bigint NOT NULL REFERENCES logins(id),
                                         id          text   NOT NULL,                         -- ID записи, задаётся клиентом
                                         login       text   NOT NULL,                         -- логин для этого ресурса
                                         password    text   NOT NULL,                         -- пароль хранится в зашифрованном виде
                                         meta        text,                                    -- произвольное описание/мета
                                         created_at  timestamptz NOT NULL DEFAULT now(),
                                         updated_at  timestamptz NOT NULL DEFAULT now(),
                                         is_deleted  bool        NOT NULL DEFAULT false,
                                         PRIMARY KEY (user_id, id)
);
CREATE INDEX IF NOT EXISTS ix_passwords_user_id ON passwords(user_id);
CREATE INDEX IF NOT EXISTS ix_passwords_user_id_updated_at
    ON passwords(user_id, updated_at);

CREATE TABLE IF NOT EXISTS texts (
                                     user_id     bigint NOT NULL REFERENCES logins(id),
                                     id          text   NOT NULL,                         -- заголовок/ID записи, задаётся клиентом
                                     text        text   NOT NULL,                         -- тело текста хранится в зашифрованном виде
                                     meta        text,                                    -- произвольное описание/мета
                                     created_at  timestamptz NOT NULL DEFAULT now(),
                                     updated_at  timestamptz NOT NULL DEFAULT now(),
                                     is_deleted  bool        NOT NULL DEFAULT false,
                                     PRIMARY KEY (user_id, id)
);
CREATE INDEX IF NOT EXISTS ix_texts_user_id ON texts(user_id);
CREATE INDEX IF NOT EXISTS ix_texts_user_id_updated_at
    ON texts(user_id, updated_at);

CREATE TABLE IF NOT EXISTS cards (
                                     user_id     bigint NOT NULL REFERENCES logins(id),
                                     card_pan    text   NOT NULL,                         -- маскированный PAN, например 4600********5363
                                     data        text   NOT NULL,                         -- зашифрованные данные карты (номер, holder, expire, cvv)
                                     meta        text,                                    -- произвольное описание/мета
                                     created_at  timestamptz NOT NULL DEFAULT now(),
                                     updated_at  timestamptz NOT NULL DEFAULT now(),
                                     is_deleted  bool        NOT NULL DEFAULT false,
                                     PRIMARY KEY (user_id, card_pan)
);
CREATE INDEX IF NOT EXISTS ix_cards_user_id ON cards(user_id);
CREATE INDEX IF NOT EXISTS ix_cards_user_id_updated_at
    ON cards(user_id, updated_at);

CREATE TABLE IF NOT EXISTS binaries (
                                        user_id     bigint NOT NULL REFERENCES logins(id),
                                        id          text   NOT NULL,                         -- ID записи, задаётся клиентом
                                        data        text   NOT NULL,                         -- зашифрованные бинарные данные (например, base64)
                                        meta        text,                                    -- произвольное описание/мета
                                        created_at  timestamptz NOT NULL DEFAULT now(),
                                        updated_at  timestamptz NOT NULL DEFAULT now(),
                                        is_deleted  bool        NOT NULL DEFAULT false,
                                        PRIMARY KEY (user_id, id)
);
CREATE INDEX IF NOT EXISTS ix_binaries_user_id ON binaries(user_id);
CREATE INDEX IF NOT EXISTS ix_binaries_user_id_updated_at
    ON binaries(user_id, updated_at);
