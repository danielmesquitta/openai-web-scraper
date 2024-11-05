.PHONY: default install update setup run lint

include .env

default: run

install:
	@go mod download && ./bin/install.sh
update:
	@go mod tidy && go get -u ./...
generate:
	@go generate ./...
setup:
	@$(CHROME_PATH) --remote-debugging-port=$(CDP_PORT) --profile-directory=Default
run:
	@go run cmd/curiosity/main.go
lint:
	@golangci-lint run && nilaway ./...
