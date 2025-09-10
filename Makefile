SERVER_BIN := jobserver
CLI_BIN := jobctl

SERVER_ENTRYPOINT := ./cmd/server
CLI_ENTRYPOINT := ./cmd/client

.PHONY: all server cli clean test
all: server cli

server:
	go build -o $(SERVER_BIN) $(SERVER_ENTRYPOINT)

cli:
	go build -o $(CLI_BIN) $(CLI_ENTRYPOINT)

clean:
	rm -f $(SERVER_BIN) $(CLI_BIN)

test:
	go test -race ./pkg/...