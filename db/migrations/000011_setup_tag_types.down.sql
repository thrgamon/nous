DROP INDEX idx_uniq_note_tag;
DROP INDEX idx_uniq_tag;
DROP TABLE notetags;
ALTER TABLE tags DROP COLUMN type;
