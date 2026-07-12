build:
	@docker compose build

run: build
	@docker compose up

restart:
	@docker compose down -v

stop:
	@docker compose stop

test:
	@go test -v ./..


