# Contributing

Thank you for considering contributing to ng-openapi-gen!

## Getting Started

1. Fork the repository
2. Clone your fork
3. Run `make build` to verify the project builds

## Development Workflow

```bash
make build    # Build the binary
make test     # Run tests with coverage
make vet      # Run go vet
make test-race # Run tests with race detection
```

### Code structure

- `cmd/ng-openapi-gen/` — CLI entrypoint
- `internal/config/` — Configuration loading and validation
- `internal/gen/` — Naming conventions and type mapping helpers
- `internal/generate/` — Code generation orchestration
- `internal/model/` — TypeScript model generation
- `internal/openapi/` — OpenAPI spec parsing
- `internal/operation/` — API operation generation
- `internal/security/` — Security scheme handling
- `internal/service/` — Angular service generation
- `templates/` — Go text/templates for code generation
- `test/` — Test fixtures (OpenAPI specs + configs)
- `out/` — Golden test baselines (expected generated output)

## Testing

- Add test fixtures in `test/` (OpenAPI spec + config)
- Generate expected output with the tool and place it in `out/`
- The `TestGoldenFiles` test compares generated output against `out/` baselines

## Pull Request Process

1. Ensure tests pass: `make test`
2. Run `go vet ./...` for code quality
3. Update documentation if adding features
4. Add test coverage for new functionality

## Reporting Issues

Report bugs and suggest features via [GitHub Issues](https://github.com/czemar/ng-openapi-gen/issues).
