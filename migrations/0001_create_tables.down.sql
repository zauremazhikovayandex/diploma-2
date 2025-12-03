DROP INDEX IF EXISTS ux_logins_login;
DROP TABLE IF EXISTS logins;

DROP INDEX IF EXISTS ix_passwords_user_id_updated_at;
DROP INDEX IF EXISTS ix_passwords_user_id;
DROP TABLE IF EXISTS passwords;

DROP INDEX IF EXISTS ix_texts_user_id_updated_at;
DROP INDEX IF EXISTS ix_texts_user_id;
DROP TABLE IF EXISTS texts;

DROP INDEX IF EXISTS ix_cards_user_id_updated_at;
DROP INDEX IF EXISTS ix_cards_user_id;
DROP TABLE IF EXISTS cards;

DROP INDEX IF EXISTS ix_binaries_user_id_updated_at;
DROP INDEX IF EXISTS ix_binaries_user_id;
DROP TABLE IF EXISTS binaries;
