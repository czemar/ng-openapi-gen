---
layout: default
title: CLI Usage
nav_order: 4
---

# CLI Usage

{: .no_toc }

All configuration options can be passed as CLI flags. Flags override values from the config file.

## Table of contents

{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Basic usage

```bash
ng-openapi-gen -i openapi.yaml -o src/app/api
```

## Flags

| Flag | Alias | Description |
|---|---|---|
| `--config` | `-c` | Path to configuration JSON file |
| `--input` | `-i` | OpenAPI spec file or URL |
| `--output` | `-o` | Output directory |
| `--silent` | | Suppress verbose output |
| `--help` | | Show help |

All configuration properties can also be passed as flags using camelCase or kebab-case:

```bash
ng-openapi-gen --input spec.yaml --model-prefix "My" --enum-style upper --promises false
```

## Examples

### Basic generation

```bash
ng-openapi-gen -i petstore.yaml -o src/app/api
```

### With URL input

```bash
ng-openapi-gen -i https://example.com/openapi.json -o src/app/api
```

### Using a config file

```bash
ng-openapi-gen -c my-config.json
```

### Multi-API setup

```bash
ng-openapi-gen -c api-users.json
ng-openapi-gen -c api-orders.json
```

---

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error (parsing, generation, etc.) |

## Environment

No environment variables are required. The tool respects `$GOPATH` if set for Go-based installs, but the binary itself is standalone.
