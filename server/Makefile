BINARY_NAME = kubin-server

.PHONY: build
build:
	go build -o dist/$(BINARY_NAME) ./cmd/server/main.go

.PHONY: clean
clean:
	rm -rf dist

.PHONY: test
test:
	go test ./...
