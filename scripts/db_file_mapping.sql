CREATE TABLE IF NOT EXISTS file_mapping (
   filename TEXT NOT NULL,
   ext TEXT NOT NULL,
   tim BIGINT,
   no BIGINT NOT NULL,
   board TEXT NOT NULL,
   PRIMARY KEY (no, board)
);

CREATE INDEX filemapping_index_tim ON file_mapping (tim);
CREATE INDEX filemapping_index_no ON file_mapping (no);