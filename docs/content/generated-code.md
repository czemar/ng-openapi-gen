---
title: Generated Code
weight: 5
---

ng-openapi-gen produces TypeScript files organized into functional groups. This page explains the output structure and how to use each piece.

## Output structure

```
src/app/api/
├── api-configuration.ts    # Injectable config (rootUrl, etc.)
├── strict-http-response.ts # Typed response wrapper
├── request-builder.ts      # HTTP request builder
├── api.ts                  # Main @Injectable Api service
├── functions.ts            # Re-exports all functions
├── models.ts               # Re-exports all models
├── services.ts             # Re-exports all services (if enabled)
├── index.ts                # Barrel export (if indexFile is true)
│
├── models/
│   ├── pet.ts              # Model interface
│   ├── pets.ts             # Array type alias
│   └── error.ts            # Error model
│
├── fn/
│   ├── index.ts            # Operations index
│   └── pets/
│       ├── list-pets.ts    # Function for GET /pets
│       └── create-pets.ts  # Function for POST /pets
│
└── services/               # (if services: true)
    ├── pets.service.ts     # @Injectable per tag
    └── index.ts            # Services index
```

---

## Core files

### `ApiConfiguration`

The root URL for API requests. Configure it in your Angular app providers:

```typescript
import { provideApiConfiguration } from './api/api-configuration';

export const appConfig: ApplicationConfig = {
  providers: [
    provideApiConfiguration('https://api.example.com'),
  ],
};
```

Or set it at runtime:

```typescript
const config = inject(ApiConfiguration);
config.rootUrl = 'https://api.example.com';
```

### `Api` (generated if `apiService` is not `false`)

The main service that invokes API functions:

```typescript
import { inject } from '@angular/core';
import { Api } from './api/api';
import { listPets } from './api/fn/pets/list-pets';

const api = inject(Api);
const pets = await api.invoke(listPets, { limit: 10 });

// Access full HTTP response:
const resp = await api.invoke$Response(listPets, { limit: 10 });
console.log(resp.headers.get('X-Rate-Limit'));
```

### `StrictHttpResponse`

A typed wrapper around `HttpResponse<T>`:

```typescript
interface StrictHttpResponse<T> {
  body: T;
  headers: HttpHeaders;
  status: number;
  statusText: string;
  url: string;
  ok: boolean;
  type: HttpEventType.Response;
}
```

### `RequestBuilder`

Builds Angular `HttpRequest` objects with typed parameters. Used internally by generated functions.

---

## Models

Each schema in `components/schemas` becomes a TypeScript interface or type alias:

```typescript
// Object schema -> Interface
export interface Pet {
  id: number;
  name: string;
  tag?: string;
}

// Array schema -> Type alias
export type Pets = Array<Pet>;

// Enum schema -> Type alias or enum (depending on enumStyle)
export type Flavor = 'vanilla' | 'chocolate' | 'strawberry';
```

---

## Functions

Each operation generates a standalone function:

```typescript
// Parameters interface
export interface ListPets$Params {
  limit?: number;
}

// Response type
export type ListPets$Result = Pets;

// Function
export function listPets(
  http: HttpClient,
  rootUrl: string,
  params?: ListPets$Params,
  context?: HttpContext,
): Observable<StrictHttpResponse<ListPets$Result>>;
```

### Variant naming

When multiple request body or response content types exist, variants are generated:

| Content type | Variant suffix |
|---|---|
| `application/json` | `$Json` (default, no suffix with `skipJsonSuffix`) |
| `application/xml` | `$Xml` |
| `text/plain` | `$Text` |

---

## Services (optional)

When `services: true`, an `@Injectable` class is generated per API tag:

```typescript
@Injectable({ providedIn: 'root' })
export class PetsService {
  constructor(private api: Api) {}

  listPets(params?: ListPets$Params): Promise<Pets> {
    return this.api.invoke(listPets, params);
  }
}
```

---

## Headers and interceptors

Use standard Angular `HttpInterceptorFn` for auth headers, logging, etc.:

```typescript
import { HttpInterceptorFn } from '@angular/common/http';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const clone = req.clone({
    headers: req.headers.set('Authorization', `Bearer ${token}`),
  });
  return next(clone);
};
```
