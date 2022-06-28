#!/bin/zsh

migrate -path db/migrations -database "$DATABASE_URL" up $2
