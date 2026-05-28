---
layout: default
title: Installation
nav_order: 2
---

# Installation

{: .no_toc }

## Table of contents

{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Install via Go

```bash
go install github.com/czemar/ng-openapi-gen/cmd/ng-openapi-gen@latest
```

This installs the `ng-openapi-gen` binary to your `$GOPATH/bin` (or `$HOME/go/bin`).

{: .note }
Requires Go 1.21+. Verify with `go version`.

## Download pre-built binary

Download the latest release from the [releases page](https://github.com/czemar/ng-openapi-gen/releases). Binaries are available for:

- Linux (amd64)
- macOS (amd64, arm64)
- Windows (amd64)

Extract the archive and add the binary to your `PATH`.

## Build from source

```bash
git clone https://github.com/czemar/ng-openapi-gen.git
cd ng-openapi-gen
make build
```

The binary is output as `ng-openapi-gen` in the project root.

---

## Verify installation

```bash
ng-openapi-gen --help
```

You should see the list of available CLI flags.

---

## Quick start

### 1. Create a configuration file

Create `ng-openapi-gen.json` in your Angular project root:

```json
{
  "$schema": "ng-openapi-gen-schema.json",
  "input": "path/to/openapi.yaml",
  "output": "src/app/api"
}
```

### 2. Generate the API code

```bash
ng-openapi-gen
```

### 3. Import the generated `Api` service

See the [generated code guide](generated-code) for how to use the output.

---

## CI/CD integration

Add a script to your `package.json` to regenerate on each build:

```json
{
  "scripts": {
    "generate:api": "ng-openapi-gen",
    "start": "npm run generate:api && ng serve",
    "build": "npm run generate:api && ng build"
  }
}
```

For multiple APIs:

```json
{
  "scripts": {
    "generate:api": "npm run generate:api:a && npm run generate:api:b",
    "generate:api:a": "ng-openapi-gen -c api-a.json",
    "generate:api:b": "ng-openapi-gen -c api-b.json"
  }
}
```
