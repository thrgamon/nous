DROP MATERIALIZED VIEW note_search;
CREATE MATERIALIZED VIEW note_search AS
SELECT
	notes.id,
	notes.body,
	ARRAY_AGG(DISTINCT tags.tag) AS tags,
	to_tsvector(notes.body || ' '  || string_agg(tags.tag, ' ')) AS doc
FROM
	notes
	LEFT JOIN tags ON tags.note_id = notes.id
GROUP BY
	notes.id;

CREATE UNIQUE INDEX idx_unq_search ON note_search (id);
CREATE INDEX idx_fts_search ON note_search USING gin(doc);


