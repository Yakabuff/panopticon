CREATE TABLE IF NOT EXISTS thread (
   id SERIAL PRIMARY KEY,
   no BIGINT NOT NULL,
   time BIGINT NOT NULL,
   name TEXT NOT NULL,
   trip TEXT,
   sub TEXT NOT NULL,
   com TEXT NOT NULL,
   replies BIGINT NOT NULL,
   images BIGINT NOT NULL,
   board TEXT NOT NULL
);

CREATE INDEX thread_index_no ON thread (no);
CREATE INDEX thread_index_time ON thread (time);