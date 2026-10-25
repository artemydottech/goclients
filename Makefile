.PHONY: run fmt vet test cover check

run:
	go run .

fmt:
	gofmt -l .

vet:
	go vet ./...

test:
	go test -race ./...

check: fmt vet test
	go build ./...

cover:
	go test -race -cover ./...
