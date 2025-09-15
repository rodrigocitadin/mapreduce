ifneq (,$(wildcard .env))
  include ./.env
  export
endif

MASTER_ADDR ?= ":9991"

worker:
	go run ./cmd/worker/main.go --masterAddr $(MASTER_ADDR)

master:
	go run ./cmd/master/main.go --masterAddr $(MASTER_ADDR)

test:
	go test ./pkg/*
