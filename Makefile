.PHONY: setconfig run lint

run: setconfig
	go build -o searchagent

setconfig:
	find config.toml &>/dev/null || cp config.example.toml config.toml

lint:
	golangci-lint run -c .golangci.yml ./...
