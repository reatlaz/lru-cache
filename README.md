# LRU cache service

#### `make build`

Build the application in bin/lru

#### `make test`

Runs the tests for the application using the command
```
go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./... && go tool cover -func coverage.out | grep total | awk '{print $3}'
```

#### `make lint`

Runs the linter on the codebase. It vendors the dependencies first and then runs `golangci-lint` with the configuration file located in `build/package/ci`.
