CREATE TABLE image_backlog (
  id BIGSERIAL PRIMARY KEY,
  board TEXT NOT NULL,
  no BIGINT NOT NULL,
  ext TEXT,
  UNIQUE(board, no)
);