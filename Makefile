.PHONY: server
server:
	go run ./cmd/server

.PHONY: server-file
server-file:
	FILE_STORAGE_PATH=./tmp/temporary.json \
	STORE_INTERVAL=300 \
	go run ./cmd/server

.PHONY: server-db
server-db:
	DATABASE_DSN="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" \
	go run ./cmd/server

.PHONE: server-db-hash
server-db-hash:
	DATABASE_DSN="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" \
	KEY="test" \
    go run ./cmd/server

.PHONE: agent-hash
agent-hash:
	KEY="test" go run ./cmd/agent

.PHONY: agent
agent:
	go run ./cmd/agent

.PHONY: migrate-up
migrate-up:
	goose -dir migrations postgres \
	"postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" up

.PHONY: migrate-down
migrate-down:
	goose -dir migrations postgres \
	"postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" down

.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres \
	"postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" status

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run