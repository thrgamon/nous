## Database
createdb learning_rank_dev
migrate create -ext sql -dir db/migrations -seq create_resource_table
migrate -path db/migrations -database "$DATABASE_URL" up
