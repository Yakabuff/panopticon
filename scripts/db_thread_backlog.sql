CREATE TABLE thread_backlog (
  id BIGSERIAL PRIMARY KEY,
  board TEXT NOT NULL,
  no BIGINT NOT NULL,
  last_modified BIGINT NOT NULL,
  last_archived BIGINT NOT NULL,
  replies BIGINT NOT NULL,
  page BIGINT,
  tid TEXT NOT NULL,
  UNIQUE(board, no)
);