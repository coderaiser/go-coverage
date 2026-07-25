.PHONY: test coverage build clean

BINARY   := coverage
CMD      := ./cmd/coverage
DIST     := dist

test:
	go test ./...

# run.go (CLI glue: os.Open, os.Args) is intentionally excluded from coverage
# by scoping -coverpkg to only the library package.
coverage:
	go test \
		-coverprofile=coverage.out \
		-covermode=atomic \
		-coverpkg=example.com/coverage \
		./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf $(DIST)
