.PHONY: build
build:
	go mod vendor
	go build -mod=vendor -o bin/lru ./cmd/lru

.PHONY: test
test:
	go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./... && go tool cover -func coverage.out | grep total | awk '{print $3}' 

.PHONY: lint
lint:
	go mod vendor
	golangci-lint run -c .golangci.yml -v --modules-download-mode=vendor ./...
