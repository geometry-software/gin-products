GO := $(if $(wildcard .tools/go/bin/go),$(CURDIR)/.tools/go/bin/go,go)
export GOPATH := $(CURDIR)/.cache/gopath
export GOCACHE := $(CURDIR)/.cache/go-build
SERVICES := adapters providers login products invoices shipping
.PHONY: deps build start vet fmt check docker-up docker-local docker-down

deps:
	$(GO) mod download

build:
	@mkdir -p bin
	@set -e; for service in $(SERVICES); do $(GO) build -trimpath -o bin/$$service ./apps/$$service; done

start: build
	bash scripts/dev.sh

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

check: vet build

docker-up:
	docker compose up --build -d --wait

docker-local:
	docker compose -f compose.yaml -f compose.local.yaml up --build -d --wait

docker-down:
	docker compose -f compose.yaml -f compose.local.yaml down
