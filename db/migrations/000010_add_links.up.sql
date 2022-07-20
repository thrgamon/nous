CREATE TABLE "links" (
  "id" SERIAL PRIMARY KEY,
  "url" text,
  "url_hash" text GENERATED ALWAYS AS (md5(url)) STORED,
  "title" text,
  "archive_status" smallint DEFAULT 1,
  "archive_job_id" text,
  "archive_exception" text,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_url ON links (url_hash);
