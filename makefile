.PHONY: check
check:
ifeq ($(OS),Windows_NT)
	go test ./...
else
	@wget -O lint-project.sh https://raw.githubusercontent.com/moov-io/infra/master/go/lint-project.sh
	@chmod +x ./lint-project.sh
	COVER_THRESHOLD=0.0 ./lint-project.sh
endif

.PHONY: setup
setup:
	docker compose up -d

.PHONY: teardown
teardown:
	docker-compose kill && docker-compose rm -f -v

.PHONY: db
db:
	psql "postgres://community_commits:secret@127.0.0.1:5432/community_commits"

.PHONY: run-example
run-example:
	go run ./cmd/community_commits -config docs/examples/config.yaml

.PHONY: clean
clean:
	@rm -rf ./bin/ ./tmp/ coverage.txt misspell* staticcheck lint-project.sh

.PHONY: cover-test cover-web
cover-test:
	go test -coverprofile=cover.out ./...
cover-web:
	go tool cover -html=cover.out
