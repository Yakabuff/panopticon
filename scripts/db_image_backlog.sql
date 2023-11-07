CREATE TABLE media_backlog (
  id BIGSERIAL PRIMARY KEY,
  board TEXT NOT NULL,
  file TEXT NOT NULL,
  date_added BIGINT NOT NULL,
  UNIQUE(board, file)
);
-- file is file identifier.. not file name