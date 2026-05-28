---
layout: default
title: Petstore example
parent: Examples
nav_order: 1
---

# Petstore example

{: .no_toc }

This walkthrough uses the [Swagger Petstore](https://petstore.swagger.io/) OpenAPI 3.0 specification to demonstrate a complete generation and usage workflow.

## Table of contents

{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Specification

The Petstore spec (saved as `petstore.yaml`) defines:

- **GET /pets** -- List all pets (with optional `limit` query param)
- **POST /pets** -- Create a new pet
- **GET /pets/{petId}** -- Find pet by ID

And schemas:
- `Pet` -- `{ id: integer, name: string, tag?: string }`
- `Pets` -- `Array<Pet>`
- `Error` -- `{ code: integer, message: string }`

---

## Configuration

Create `ng-openapi-gen.json`:

```json
{
  "$schema": "ng-openapi-gen-schema.json",
  "input": "petstore.yaml",
  "output": "src/app/api",
  "modelPrefix": "",
  "modelSuffix": "",
  "ignoreUnusedModels": false,
  "enumStyle": "alias"
}
```

Run:

```bash
ng-openapi-gen
```

---

## Generated output

After generation, your project has:

```
src/app/api/
├── api-configuration.ts
├── strict-http-response.ts
├── request-builder.ts
├── api.ts
├── functions.ts
├── models.ts
├── models/
│   ├── pet.ts          # export interface Pet { ... }
│   ├── pets.ts         # export type Pets = Array<Pet>;
│   └── error.ts        # export interface Error { ... }
└── fn/
    └── pets/
        ├── list-pets.ts
        ├── create-pets.ts
        └── show-pet-by-id.ts
```

### Model: `pet.ts`

```typescript
export interface Pet {
  id: number;
  name: string;
  tag?: string;
}
```

### Function: `list-pets.ts`

```typescript
export interface ListPets$Params {
  limit?: number;
}

export function listPets(
  http: HttpClient,
  rootUrl: string,
  params?: ListPets$Params,
  context?: HttpContext,
): Observable<StrictHttpResponse<Pets>>;
```

---

## Usage

### Setup providers

In `app.config.ts`:

```typescript
import { ApplicationConfig } from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import { provideApiConfiguration } from './api/api-configuration';

export const appConfig: ApplicationConfig = {
  providers: [
    provideHttpClient(),
    provideApiConfiguration('https://petstore.example.com/api'),
  ],
};
```

### Functional API (default)

```typescript
import { Component, inject, OnInit, signal } from '@angular/core';
import { Api } from './api/api';
import { listPets } from './api/fn/pets/list-pets';
import { Pet } from './api/models';

@Component({
  selector: 'app-pet-list',
  template: `
    <ul>
      @for (pet of pets(); track pet.id) {
        <li>{{ pet.name }}</li>
      }
    </ul>
  `,
})
export class PetListComponent implements OnInit {
  private api = inject(Api);
  readonly pets = signal<Pet[]>([]);

  async ngOnInit() {
    const result = await this.api.invoke(listPets, { limit: 20 });
    this.pets.set(result);
  }
}
```

### Service-based API (if `services: true`)

```typescript
import { Component, inject, OnInit, signal } from '@angular/core';
import { PetsService } from './api/services';
import { Pet } from './api/models';

@Component({ ... })
export class PetListComponent implements OnInit {
  private petsService = inject(PetsService);
  readonly pets = signal<Pet[]>([]);

  async ngOnInit() {
    const result = await this.petsService.listPets({ limit: 20 });
    this.pets.set(result);
  }
}
```

### Full response access

```typescript
const resp = await this.api.invoke$Response(listPets, { limit: 20 });
console.log('Status:', resp.status);
console.log('Headers:', resp.headers.get('X-Total-Count'));
```

---

## Error handling

Wrap API calls in try/catch:

```typescript
try {
  const pets = await this.api.invoke(listPets, { limit: 20 });
} catch (err) {
  console.error('API Error:', err);
}
```

Or use an `HttpInterceptorFn` for global error handling:

```typescript
import { HttpInterceptorFn } from '@angular/common/http';
import { catchError } from 'rxjs/operators';
import { throwError } from 'rxjs';

export const errorInterceptor: HttpInterceptorFn = (req, next) =>
  next(req).pipe(
    catchError((error) => {
      console.error('HTTP Error:', error);
      return throwError(() => error);
    }),
  );
```
