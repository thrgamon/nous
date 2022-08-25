DROP MATERIALIZED VIEW note_search;
CREATE MATERIALIZED VIEW note_search AS
SELECT notes.id,
  array_agg(COALESCE(tags.tag, '')) AS tags,
  to_tsvector((notes.body || ' ')) || setweight(to_tsvector(string_agg(COALESCE(tags.tag, ''), ' ')), 'A') AS doc
 FROM notes
   JOIN notetags on notes.id = notetags.note_id
   JOIN tags ON notetags.tag_id = tags.id
GROUP BY notes.id;

CREATE UNIQUE INDEX idx_unq_search ON note_search (id);
CREATE INDEX idx_fts_search ON note_search USING gin(doc);
