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
benchmark:
	go test -bench=. -benchmem ./...
.PHONY: build run clean test
