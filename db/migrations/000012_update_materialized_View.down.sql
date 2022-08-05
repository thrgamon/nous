DROP TRIGGER refresh_note_search on notetags;

CREATE TRIGGER refresh_note_search
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON tags
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_note_search();

DROP MATERIALIZED VIEW note_search;
CREATE MATERIALIZED VIEW note_search AS
SELECT notes.id,
  notes.body,
  array_agg(DISTINCT COALESCE(tags.tag, ''::character varying)) AS tags,
  notes.done,
  notes.inserted_at,
  to_tsvector((notes.body || ' '::text)) || setweight(to_tsvector(string_agg(COALESCE(tags.tag, ''::character varying)::text, ' '::text)), 'A') AS doc
 FROM notes
   LEFT JOIN tags ON tags.note_id = notes.id
GROUP BY notes.id;

CREATE UNIQUE INDEX idx_unq_search ON note_search (id);
CREATE INDEX idx_fts_search ON note_search USING gin(doc);



