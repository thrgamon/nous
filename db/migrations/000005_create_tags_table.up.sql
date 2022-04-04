CREATE TABLE "tags" (
  "id" SERIAL PRIMARY KEY,
  "tag" varchar,
  "user_id" int,
  "resource_id" int,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_resource FOREIGN KEY(resource_id) REFERENCES resources(id)
);

CREATE INDEX idx_resource_tag ON tags (resource_id, tag);

