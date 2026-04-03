# transport/http/dto

Data Transfer Objects for the HTTP layer. These structs are the only types in the service that carry `json:` struct tags.

## Files

| File | Purpose |
|---|---|
| `request.go` | Inbound DTOs parsed from query parameters or JSON request bodies. Each DTO has a `Validate()` method and a `ToDomain()` method that converts it to the appropriate domain type. |
| `response.go` | Outbound DTOs serialised to JSON in HTTP responses. Each has a `FromDomain*` constructor that maps from a domain type. |

## Why DTOs are separate from domain models

Domain structs represent business concepts and must be free of serialisation concerns. Keeping DTOs here means:

- `json:` tags, `omitempty`, and field name choices are isolated to this package
- The API contract (field names, response shape) can evolve independently of the domain model
- Domain structs can be renamed or restructured without breaking JSON responses, and vice versa

## Validation

Request DTOs validate their own fields in `Validate() error`. Handlers call this immediately after parsing and return a 400 before reaching the use-case if validation fails. Validation here covers only shape constraints (required fields, coordinate ranges) — business-rule validation happens in the domain or application layer.
