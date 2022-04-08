## Database
Just remembering some commands I always forget.
`createdb learning_rank_dev`
`migrate create -ext sql -dir db/migrations -seq create_resource_table`
`migrate -path db/migrations -database "$DATABASE_URL" up`

## Todo
- Search
- Better error handling
- Tests
