.PHONY: test test-race test-coverage test-watch test-usecase lint

PKGS := $(shell go list ./... | grep -v /e2e)

test:
	go test $(PKGS)

test-race:
	go test -race $(PKGS)

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic -coverpkg=./pkg/... $(PKGS)
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "→ open coverage.html"

test-usecase:
	go test -race -v ./pkg/usecase/...

test-watch:
	ginkgo watch -r ./pkg

lint:
	golangci-lint run
