.PHONY: all build test test-race vet clean

BINARY=ng-openapi-gen
COVERPROFILE=coverage.out

all: test vet build

build:
	go build -o $(BINARY) ./cmd/$(BINARY)/

test:
	go test -coverprofile=$(COVERPROFILE) ./...
	go tool cover -func=$(COVERPROFILE)
	go tool cover -html=$(COVERPROFILE) -o coverage.html

test-race:
	go test -race ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(COVERPROFILE) coverage.html
