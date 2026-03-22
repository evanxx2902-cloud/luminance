.PHONY: generate build

generate:
	go run entgo.io/ent/cmd/ent generate ./ent/schema

build:
	CGO_ENABLED=0 GOOS=linux go build -o luminance-api ./cmd/luminance-api
	CGO_ENABLED=0 GOOS=linux go build -o luminance-migrate ./cmd/migrate
