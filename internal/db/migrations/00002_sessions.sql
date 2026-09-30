-- +goose Up
-- 로그인 세션. 쿠키에는 무작위 토큰 원문, DB에는 그 SHA-256 해시만 저장한다.
-- DB가 유출돼도 해시로는 쿠키를 만들 수 없다.
CREATE TABLE sessions (
    token_hash TEXT     PRIMARY KEY,
    user_id    INTEGER  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sessions_user ON sessions (user_id);
CREATE INDEX idx_sessions_expires ON sessions (expires_at);

-- +goose Down
DROP TABLE sessions;
