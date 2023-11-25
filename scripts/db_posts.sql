CREATE TABLE IF NOT EXISTS post (
   id SERIAL PRIMARY KEY,
   no BIGINT NOT NULL,
   resto BIGINT NOT NULL,
   time BIGINT NOT NULL,
   name TEXT NOT NULL,
   trip TEXT,
   com TEXT NOT NULL,
   board TEXT NOT NULL,
   tid TEXT NOT NULL,
   pid TEXT NOT NULL,
   UNIQUE(tid, pid)
);

CREATE INDEX post_index_resto ON post (resto);
CREATE INDEX post_index_tid ON post (tid);
CREATE INDEX post_index_pid ON post (pid);