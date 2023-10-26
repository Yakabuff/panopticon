CREATE TABLE IF NOT EXISTS boards (
   board TEXT PRIMARY KEY,
   title TEXT NOT NULL,
   unlisted bool NOT NULL
);