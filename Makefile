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
	go generate ./data/ent

run:
	go run .

clean:
	go clean -cache
