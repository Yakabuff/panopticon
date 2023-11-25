CREATE TABLE IF NOT EXISTS file_mapping (
   filename TEXT NOT NULL,
   ext TEXT NOT NULL,
   identifier TEXT,
   no BIGINT NOT NULL,
   board TEXT NOT NULL,
   fileid BIGINT NOT NULL,
   tid TEXT,
   pid TEXT,
   PRIMARY KEY (no, board, tid, pid, identifier),
   CONSTRAINT fk_fileid
      FOREIGN KEY(fileid)
         REFERENCES file(id)
);

CREATE INDEX filemapping_index_tim ON file_mapping (tim);
CREATE INDEX filemapping_index_no ON file_mapping (no);