---
layout: home
title: Home
nav_order: 1
---

# ng-openapi-gen

{: .fs-6 }
An **OpenAPI 3.0 and 3.1 code generator** for **Angular 16+**.

{: .fs-4 }
Generates TypeScript model interfaces, typed API functions, and Angular `@Injectable` services directly from your OpenAPI specification.

---

## Why ng-openapi-gen?

- **Type-safe API clients** -- Models, parameters, and responses are fully typed
- **OpenAPI 3.0 & 3.1** -- Supports both versions in JSON and YAML
- **Promise & Observable** -- Choose your preferred async model
- **Tree-shakeable** -- Only the functions you use end up in your bundle
- **Customizable templates** -- Override any generated file with Go templates
- **Works with Angular CLI** -- Easy integration into existing projects

---

## Quick start

```bash
# Install
go install github.com/czemar/ng-openapi-gen/cmd/ng-openapi-gen@latest

# Generate code from your OpenAPI spec
ng-openapi-gen --input my-api.yaml --output src/app/api
```

Then in your Angular component:

```typescript
import { inject } from '@angular/core';
import { Api } from './api/api';
import { listPets } from './api/fn/pets/list-pets';

export class MyComponent {
  private api = inject(Api);

  async ngOnInit() {
    const pets = await this.api.invoke(listPets, { limit: 10 });
  }
}
```

---

## Key features

| Feature | Description |
|---|---|
| Functional API | Each operation is a standalone function -- only bundle what you use |
| Services (optional) | Traditional `@Injectable` services per API tag |
| Models | TypeScript interfaces for all schemas, with prefix/suffix support |
| Enums | Configurable enum styles: `alias`, `upper`, `pascal`, `ignorecase` |
| Content negotiation | Multiple request/response content types generate distinct variants |
| Promise or Observable | Defaults to Promise; switch to Observable with one flag |
| Tag filtering | Include or exclude operations by tag |
| Custom templates | Override any output file with your own Go template |
| Strict TypeScript | Generated code compiles with `noUnusedLocals`, `noUnusedParameters` |

---

## Next steps

- [Installation and setup](installation)
- [Configuration reference](configuration)
- [CLI usage](cli-usage)
- [Petstore example walkthrough](examples/petstore)
