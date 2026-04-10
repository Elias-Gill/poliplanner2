-- +migrate Up
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP TABLE IF EXISTS user_sessions;
