.DEFAULT_GOAL := help
MASTER_ADDR ?= ":9991"

.PHONY: master
master:
	@echo ">> running master for wordcount example"
	go run ./examples/wordcount/wordcount.go --role=master --masterAddr $(MASTER_ADDR)

.PHONY: wordcount-worker
wordcount-worker:
	@echo ">> running wordcount worker"
	go run ./examples/wordcount/wordcount.go --role=worker --masterAddr $(MASTER_ADDR) --workerAddr ":1234"

.PHONY: wordcount-submit
wordcount-submit:
	@echo ">> submitting wordcount job"
	go run ./examples/wordcount/wordcount.go --role=submit --masterAddr $(MASTER_ADDR) --inputs ./inputs

.PHONY: test
test:
	@echo ">> running tests"
	go test ./pkg/*

.PHONY: script
script:
	@echo ">> running wordcount example with existing inputs"
	chmod +x ./run-wordcount.sh && ./run-wordcount.sh

.PHONY: help
help:
	@echo "Makefile for MapReduce Library"
	@echo ""
	@echo "Usage:"
	@echo "  make master              - Run the master server for the wordcount example."
	@echo "  make wordcount-worker    - Run a worker for the wordcount example."
	@echo "  make wordcount-submit    - Submit the wordcount job to the master."
	@echo "  make test                - Run all go tests."
	@echo "  make script              - Run the wordcount test with 3 workers."
	@echo ""
	@echo "Typical workflow:"
	@echo "1. In one terminal, run 'make master'."
	@echo "2. In one or more other terminals, run 'make wordcount-worker'."
	@echo "3. In another terminal, run 'make wordcount-submit' to start the job."
