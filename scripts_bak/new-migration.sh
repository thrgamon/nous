#!/bin/zsh

migrate create -ext sql -dir db/migrations -seq $1
