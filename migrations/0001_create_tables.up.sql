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
