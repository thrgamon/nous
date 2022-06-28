#!/bin/zsh

migrate -path db/migrations -database "$DATABASE_URL" down $2
