# Changelog

## [Unreleased]

### Added
- Golden test -update flag for regenerating baselines (`go test -update ./internal/generate/`)
- golangci-lint configuration
- Dependabot config for automated dependency updates
- Package-level doc comments on all internal packages
- Test coverage for `security/` and `cmd/ng-openapi-gen/` packages
- Shared test helpers in `internal/testutil/`
- CODE_OF_CONDUCT.md, CONTRIBUTING.md, SECURITY.md
- GitHub issue and PR templates

### Fixed
- `out/` no longer gitignored — golden test baselines are now tracked in git
- Missing `keep-model-names.json` and `self-ref-array.json` test fixtures (golden test was silently skipping these cases)
- Duplicated test helpers consolidated across 6 packages
- `.gitignore` now covers `.DS_Store`, `Thumbs.db`, `*.exe`, `*.test`, `tmp/`
