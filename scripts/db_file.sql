CREATE TABLE IF NOT EXISTS file (
   id SERIAL PRIMARY KEY,
   sha256 TEXT NOT NULL,
   md5 TEXT NOT NULL,
   w INT NOT NULL,
   h INT NOT NULL,
   fsize INT NOT NULL,
   mime TEXT NOT NULL,
   UNIQUE(sha256, md5, w, h, fsize, mime)
);

CREATE INDEX file_index_sha2 ON file (sha256);
CREATE INDEX file_index_md5 ON file (md5);