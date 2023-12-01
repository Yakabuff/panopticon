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
   board TEXT NOT NULL,
   tid TEXT UNIQUE NOT NULL,
   has_image boolean NOT NULL,
);

CREATE INDEX thread_index_tid ON thread (tid);
CREATE INDEX thread_index_time ON thread (time);