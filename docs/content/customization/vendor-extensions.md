---
title: Vendor extensions
weight: 2
---

ng-openapi-gen supports these OpenAPI vendor extensions.

---

## `x-operation-name`

LoopBack compatible. Override the generated method name for an operation. Useful when you want shorter or more domain-specific names per tag.

```yaml
paths:
  /users:
    get:
      tags:
        - Users
      operationId: listUsers
      x-operation-name: list
  /places:
    get:
      tags:
        - Places
      operationId: listPlaces
      x-operation-name: list
```

Result:
- `UsersService.list()` instead of `UsersService.listUsers()`
- `PlacesService.list()` instead of `PlacesService.listPlaces()`

---

## `x-enumNames`

NSwag compatible. Customize TypeScript enum member names. Must be an array with the same length as the enum values.

```yaml
components:
  schemas:
    HttpStatusCode:
      type: integer
      enum:
        - 200
        - 404
        - 500
      x-enumNames:
        - OK
        - NOT_FOUND
        - INTERNAL_SERVER_ERROR
```

Result (with `enumStyle: "alias"`):

```typescript
export type HttpStatusCode = 200 | 404 | 500;
export const HttpStatusCodeValues = {
  OK: 200,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
} as const;
```
