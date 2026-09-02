BINARY_NAME=narvi
DB_URL=postgres://postgres:qwerty@localhost:5432/narvi?sslmode=disable
PWD := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
REPORT_DIR := tests/reports/sast

.PHONY: lint
.PHONY: clean
.PHONY: run
.PHONY: test
.PHONY: build

lint:
	golangci-lint run --fix --config=.golangci.yaml

sast-semgrep:
	mkdir -p $(REPORT_DIR)
	docker run --rm -v "$(PWD)":/src -w /src semgrep/semgrep semgrep --config=auto --text > $(REPORT_DIR)/semgrep.txt
	echo "SAST semgrep completed"

run:
	go run ./cmd/main.go -config=resources/config.toml

test:
	go test -v ./... -count=1

build:
	CGO_ENABLED=0 go build -o ./out/${BINARY_NAME} ./cmd/main.go

clean:
	go clean
	rm -f ./out

docker-build:
	docker build -t ${BINARY_NAME} .

migrations-up:
	dbmate -u "${DB_URL}" up
migrations-down:
	dbmate -u "${DB_URL}" down
migrations-info:
	dbmate -u "${DB_URL}" status
migrations-drop:
	dbmate -u "${DB_URL}" drop

generate-ed25519:
	openssl genpkey -algorithm ed25519 | base64

generate-hmac-key:
	openssl rand -base64 32
