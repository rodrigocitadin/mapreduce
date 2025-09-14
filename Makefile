ifneq (,$(wildcard .env))
  include ./.env
  export
endif

MASTER_ADDR ?= ":9991"
WORKER_ADDR ?= ":9992"

worker:
	go run ./cmd/worker/main.go --masterAddr $(MASTER_ADDR) --workerAddr $(WORKER_ADDR)

master:
	go run ./cmd/master/main.go --masterAddr $(MASTER_ADDR)
