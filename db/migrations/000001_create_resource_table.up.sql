CREATE TABLE "resources" (
  "id" SERIAL PRIMARY KEY,
  "link" text,
  "name" text,
  "rank" int,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
