DROP MATERIALIZED VIEW note_search;
CREATE MATERIALIZED VIEW note_search AS
SELECT notes.id,
  notes.body,
  array_agg(COALESCE(tags.tag, '')) AS tags,
  notes.done,
  notes.inserted_at,
  to_tsvector((notes.body || ' ')) || setweight(to_tsvector(string_agg(COALESCE(tags.tag, ''), ' ')), 'A') AS doc,
  notes.reviewed_at
 FROM notes
   LEFT JOIN notetags on notes.id = notetags.note_id
   LEFT JOIN tags ON notetags.tag_id = tags.id
GROUP BY notes.id;

