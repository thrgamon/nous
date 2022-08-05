ALTER TABLE tags ADD type smallint DEFAULT 1 NOT NULL;

CREATE TABLE "notetags" (
  "tag_id" int NOT NULL,
  "note_id" int NOT NULL,
  CONSTRAINT fk_tag FOREIGN KEY(tag_id) REFERENCES tags(id),
  CONSTRAINT fk_note FOREIGN KEY(note_id) REFERENCES notes(id)
);

CREATE UNIQUE INDEX idx_uniq_note_tag ON notetags (note_id, tag_id);
CREATE UNIQUE INDEX idx_uniq_tag ON tags (tag);

