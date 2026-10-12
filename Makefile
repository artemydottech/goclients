.PHONY: run fmt vet test check

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
