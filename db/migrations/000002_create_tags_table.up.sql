CREATE TABLE "tags" (
  "id" SERIAL PRIMARY KEY,
  "tag" varchar,
  "note_id" int,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT fk_note FOREIGN KEY(note_id) REFERENCES notes(id)
);

CREATE INDEX idx_note_tag ON tags (note_id, tag);

