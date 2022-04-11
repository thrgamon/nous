CREATE TABLE "comments" (
  "id" SERIAL PRIMARY KEY,
  "content" text,
  "user_id" int,
  "resource_id" int,
  "parent_id" int,
  "inserted_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
  CONSTRAINT fk_resource FOREIGN KEY(resource_id) REFERENCES resources(id),
  CONSTRAINT fk_comment_parent FOREIGN KEY(parent_id) REFERENCES comments(id)
);

CREATE INDEX idx_fk_comments_resource ON comments (resource_id);
CREATE INDEX idx_fk_comments_user ON comments (user_id);
CREATE INDEX idx_fk_comments_parent ON comments (parent_id);
