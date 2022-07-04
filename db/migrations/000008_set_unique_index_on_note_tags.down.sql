DROP INDEX idx_note_tag;
CREATE INDEX idx_note_tag ON tags (note_id, tag);


