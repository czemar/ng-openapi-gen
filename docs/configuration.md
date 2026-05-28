---
layout: default
title: Configuration
nav_order: 3

---

# Configuration

{: .no_toc }

ng-openapi-gen is configured via a JSON file (default `ng-openapi-gen.json`) or CLI flags. The only required property is `input`.

## Table of contents

{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Example

```json
{
  "$schema": "ng-openapi-gen-schema.json",
  "input": "my-api.yaml",
  "output": "src/app/api",
  "ignoreUnusedModels": false,
  "enumStyle": "alias",
  "promises": true
}
```

---

## Reference

### `input` {: .required }

> **Required.** The OpenAPI specification file.

- Type: `string`
- Accepts local file paths or remote URLs
- Supports JSON and YAML formats

### `output`

> Where generated files are written.

- Type: `string`
- Default: `"src/app/api"`

### `fetchTimeout`

> Timeout when fetching a remote URL, in milliseconds.

- Type: `number`
- Default: `20000` (20 seconds)

---

### `defaultTag`

> Tag name used for operations that have no tags.

- Type: `string`
- Default: `"Api"`

### `includeTags`

> Only generate code for operations with these tags.

- Type: `string[]`

### `excludeTags`

> Exclude operations with these tags from generation.

- Type: `string[]`

### `excludeParameters`

> Ignore parameters with these names in generated services.

- Type: `string[]`

---

### `configuration`

> Name of the generated configuration class.

- Type: `string`
- Default: `"ApiConfiguration"`

### `apiService`

> Name of the generated API service used to invoke functions. Set to `false` to skip.

- Type: `string | boolean`
- Default: `"Api"`

### `functionIndex`

> Name of the TypeScript file (without `.ts`) that re-exports all functions. Set to `false` to skip.

- Type: `string | boolean`
- Default: `"functions"`

### `modelIndex`

> Name of the TypeScript file (without `.ts`) that re-exports all models. Set to `false` to skip.

- Type: `string | boolean`
- Default: `"models"`

### `serviceIndex`

> Name of the TypeScript file (without `.ts`) that re-exports all services. Ignored unless `services` is `true`.

- Type: `string | boolean`
- Default: `"services"`

### `indexFile`

> Generate an `index.ts` that re-exports everything.

- Type: `boolean`
- Default: `false`

---

### `modelPrefix`

> Prefix added to all generated model class names.

- Type: `string`
- Default: `""`

### `modelSuffix`

> Suffix added to all generated model class names.

- Type: `string`
- Default: `""`

### `camelizeModelNames`

> When `true`, model names are camelized (in addition to capitalizing the first letter).

- Type: `boolean`
- Default: `true`

### `ignoreUnusedModels`

> Skip generating model files for schemas not referenced by any operation.

- Type: `boolean`
- Default: `true`

---

### `enumStyle`

> How enum models are generated.

- Type: `"alias" | "upper" | "pascal" | "ignorecase"`
- Default: `"alias"`

| Style | Output |
|---|---|
| `alias` | `type Flavor = 'vanilla' \| 'chocolate'` (zero bundle cost) |
| `upper` | `enum Flavor { VANILLA = ..., CHOCOLATE = ... }` |
| `pascal` | `enum Flavor { Vanilla = ..., Chocolate = ... }` |
| `ignorecase` | `enum Flavor { vanilla = 'vanilla', ... }` |

### `enumArray`

> Generate an array export alongside each enum for iteration.

- Type: `boolean`
- Default: `true` (when `enumStyle` is `alias`)

---

### `promises`

> When `true`, generated functions return `Promise`. When `false`, they return `Observable`.

- Type: `boolean`
- Default: `true`

### `services`

> Generate `@Injectable` service classes (one per API tag). Adds convenience at the cost of bundle size.

- Type: `boolean`
- Default: `false`

### `servicePrefix`

> Prefix for generated service class names. Ignored unless `services` is `true`.

- Type: `string`
- Default: `""`

### `serviceSuffix`

> Suffix for generated service class names. Ignored unless `services` is `true`.

- Type: `string`
- Default: `"Service"`

### `baseService`

> Name for the base service class (if services are generated).

- Type: `string`
- Default: `"BaseService"`

---

### `module`

> Name of an Angular `NgModule` class that provides all services. Set to `false` (default) to skip NgModule generation (recommended for standalone component projects).

- Type: `string | boolean`
- Default: `false`

### `requestBuilder`

> Name of the request builder class.

- Type: `string`
- Default: `"RequestBuilder"`

### `response`

> Name of the typed response wrapper class.

- Type: `string`
- Default: `"StrictHttpResponse"`

---

### `removeStaleFiles`

> Remove files in the output directory that were not generated.

- Type: `boolean`
- Default: `true`

### `useTempDir`

> Write temporary files to the system temp directory instead of the output directory.

- Type: `boolean`
- Default: `false`

### `silent`

> Suppress verbose output during generation.

- Type: `boolean`
- Default: `false`

---

### `templates`

> Path to a directory with custom Go templates. Any `.go.tmpl` files here override the built-in templates.

- Type: `string`

### `endOfLineStyle`

> Line ending normalization for generated files.

- Type: `"crlf" | "lf" | "cr" | "auto"`
- Default: `"auto"`

### `skipJsonSuffix`

> Skip the `$Json` suffix in generated method names for JSON content types.

- Type: `boolean`
- Default: `false`

### `customizedResponseType`

> Override the HTTP response type (`arraybuffer`, `blob`, `json`, `document`) for specific paths.

- Type: `object`

Example:

```json
{
  "customizedResponseType": {
    "/api/download": { "toUse": "blob" }
  }
}
```

### `keepFullResponseMediaType`

> Control how media types are abbreviated in generated method names.

- Type: `boolean | array`
- Default: `false`

When an array is given, each entry specifies a `mediaType` regex and a `use` strategy:

| Strategy | Example |
|---|---|
| `"short"` | `getEntities$Json` |
| `"tail"` | `getEntities$XSpringDataCompactJson` |
| `"full"` | `getEntities$ApplicationXSpringDataCompactJson` |

---

## JSON Schema

A complete [JSON Schema](https://github.com/czemar/ng-openapi-gen/blob/master/ng-openapi-gen-schema.json) is available for IDE autocompletion and validation.
