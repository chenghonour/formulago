.PHONY: build test lint clean run generate vet tidy

build:
	go build ./...

test:
	go test ./... -v -count=1

vet:
	go vet ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

generate:
	go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration ./data/ent/schema

run:
	go run .

clean:
	go clean -cache
