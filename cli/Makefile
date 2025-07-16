BINARY_NAME = kubin
VERSION     = 0.0.1
GOOS        = linux darwin windows
GOARCH      = amd64

.PHONY: build
build:
	go build -o dist/$(BINARY_NAME) ./main.go

.PHONY: clean
clean:
	rm -rf dist

.PHONY: release
release:
	@for os in $(GOOS); do \
		for arch in $(GOARCH); do \
			GOOS=$$os GOARCH=$$arch go build -ldflags "-X main.version=$(VERSION)" \
			-o dist/$(BINARY_NAME)-$(VERSION)-$$os-$$arch ./main.go; \
		done; \
	done

.PHONY: test
test:
	go test ./...
