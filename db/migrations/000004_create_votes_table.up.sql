CREATE TABLE "votes" (
  "id" SERIAL PRIMARY KEY,
  "user_id" int,
  "resource_id" int,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_resource FOREIGN KEY(resource_id) REFERENCES resources(id)
);

CREATE UNIQUE INDEX idx_uniq_resource_user ON votes (resource_id, user_id);
CREATE INDEX idx_user_resource ON votes (user_id, resource_id);

