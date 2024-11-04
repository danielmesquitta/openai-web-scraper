.PHONY: default install update run clear generate build lint migrate seed new_entity

include .env

default: run

install:
	@go mod download && ./bin/install.sh
update:
	@go mod tidy && go get -u ./...
setup:
	@$(CHROME_PATH) --remote-debugging-port=9222
run:
	@go run cmd/scraper/main.go
lint:
	@golangci-lint run && nilaway ./...
