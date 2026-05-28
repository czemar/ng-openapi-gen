---
layout: default
title: Custom templates
parent: Customization
nav_order: 1
---

# Custom templates

{: .no_toc }

You can override any generated file by providing custom [Go text/templates](https://pkg.go.dev/text/template).

## Table of contents

{: .no_toc .text-delta }

1. TOC
{:toc}

---

## How it works

Copy a built-in template from the `templates/` directory, customize it, and point the `templates` config option to your folder:

```json
{
  "templates": "src/templates"
}
```

Any `.go.tmpl` file in your templates directory overrides the corresponding built-in template.

---

## Available templates

| Template | Generates |
|---|---|
| `model.go.tmpl` | Model interfaces |
| `fn.go.tmpl` | Operation functions |
| `service.go.tmpl` | Service classes |
| `apiService.go.tmpl` | Main `Api` service |
| `configuration.go.tmpl` | `ApiConfiguration` class |
| `baseService.go.tmpl` | Base service class |
| `requestBuilder.go.tmpl` | Request builder |
| `response.go.tmpl` | `StrictHttpResponse` |
| `module.go.tmpl` | NgModule |
| `index.go.tmpl` | Barrel exports |
| `functionIndex.go.tmpl` | Function index |
| `modelIndex.go.tmpl` | Model index |
| `serviceIndex.go.tmpl` | Service index |
| `enumArray.go.tmpl` | Enum array export |

---

## Template functions

These functions are available inside templates in addition to Go's built-in template functions:

| Function | Description |
|---|---|
| `upperFirst` | Capitalizes the first character |
| `lowerFirst` | Lowercases the first character |
| `camelCase` | Converts to camelCase |
| `kebabCase` | Converts to kebab-case |
| `fileName` | Converts a class name to kebab-case file name |
| `typeName` | Converts to a TypeScript type name |
| `escapeJS` | Escapes a string for JavaScript |
| `escapeId` | Returns a valid JavaScript identifier |
| `tsComments` | Formats TypeScript JSDoc comments |
| `join` | Joins strings with separator |
| `add` | Integer addition |
| `seq` | Generates a sequence of integers |
| `dict` | Creates a map from key-value pairs |

---

## Example: Extending model interfaces

Copy `templates/model.go.tmpl` to `src/templates/model.go.tmpl` and edit it to extend a base interface:

{% raw %}
```gotemplate
export interface {{ .TypeName }} extends BaseModel {
  {{- range .Properties }}
  {{ tsComments .Description 1 .Deprecated }}{{ .Name }}{{ if not .Required }}?{{ end }}: {{ .Type }};
  {{- end }}
}
```
{% endraw %}

Now all generated models extend `BaseModel`:
```typescript
export interface Pet extends BaseModel {
  id: number;
  name: string;
  tag?: string;
}
```
