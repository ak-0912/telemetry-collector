APP_NAME=collector
CACHE_ENV=XDG_CACHE_HOME="$(CURDIR)/.cache" GOCACHE="$(CURDIR)/.cache/go-build" GOPATH="$(CURDIR)/.gopath" GOMODCACHE="$(CURDIR)/.gopath/pkg/mod"
PROTO_CACHE_ENV=XDG_CACHE_HOME="/tmp/telemetry-cache" GOCACHE="/tmp/telemetry-go-build" GOPATH="/tmp/telemetry-gopath" GOMODCACHE="/tmp/telemetry-gopath/pkg/mod"
BUF_CMD=github.com/bufbuild/buf/cmd/buf

.PHONY: build vet lint run stop test test-coverage proto

build: vet lint
	go build -o bin/$(APP_NAME) ./cmd/collector

vet:
	go vet ./...

lint:
	$(CACHE_ENV) go run honnef.co/go/tools/cmd/staticcheck@latest ./...

run:
	go run ./cmd/collector

stop:
	docker compose down

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

proto:
	$(PROTO_CACHE_ENV) go run $(BUF_CMD) generate --path api
