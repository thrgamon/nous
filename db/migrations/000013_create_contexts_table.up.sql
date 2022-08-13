CREATE TABLE "contexts" (
  "id" SERIAL PRIMARY KEY,
  "context" varchar(80),
  "active" bool NOT NULL DEFAULT false
);

INSERT INTO "contexts" ("context", "active") VALUES
('work', 'f'),
('home', 't');
