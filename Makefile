BINARY_NAME=go-mem-kv
build:
	go build -o $(BINARY_NAME) main.go
run:
	go run main.go
clean:
	go clean
	rm -f $(BINARY_NAME)
test:
	go test ./...
benchmark: build
	go test -bench=. -benchmem ./kvstore/... -args -timeout=${TIMEOUT}
.PHONY: build run clean test
