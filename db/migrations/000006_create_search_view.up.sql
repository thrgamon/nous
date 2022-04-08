CREATE MATERIALIZED VIEW resource_search AS
SELECT
	resources.id,
	resources.link,
	resources.name,
	COUNT(DISTINCT votes.user_id) AS rank,
	ARRAY_AGG(DISTINCT tags.tag) AS tags,
	to_tsvector(resources.name || ' ' || resources.link || ' ' || string_agg(tags.tag, ' ')) AS doc
FROM
	resources
	LEFT JOIN votes ON votes.resource_id = resources.id
	LEFT JOIN tags ON tags.resource_id = resources.id
GROUP BY
	resources.id;

CREATE UNIQUE INDEX idx_unq_search ON resource_search (id);
CREATE INDEX idx_fts_search ON resource_search USING gin(doc);

CREATE OR REPLACE FUNCTION refresh_resource_search()
RETURNS TRIGGER LANGUAGE plpgsql
AS $$
BEGIN
REFRESH MATERIALIZED VIEW CONCURRENTLY resource_search;
RETURN NULL;
END $$;

CREATE TRIGGER refresh_resource_search
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON resources
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_resource_search();

CREATE TRIGGER refresh_resource_search
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON tags
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_resource_search();

CREATE TRIGGER refresh_resource_search
AFTER INSERT OR UPDATE OR DELETE OR TRUNCATE
ON votes
FOR EACH STATEMENT
EXECUTE PROCEDURE refresh_resource_search();
