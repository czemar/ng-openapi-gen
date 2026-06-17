.PHONY: all build test test-race vet clean wasm docs lint

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

lint:
	golangci-lint run ./...

wasm:
	GOOS=js GOARCH=wasm go build -o docs/static/demo/demo.wasm ./cmd/wasm/

docs: wasm
	hugo server --source docs

clean:
	rm -f $(BINARY) $(COVERPROFILE) coverage.html
