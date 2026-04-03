# internal

All application code lives here. The `internal` package name is a Go convention that prevents any code outside this module from importing these packages — enforcing that the venue-service's internals are never used as a library by other services.

## Sub-packages

| Package | Role |
|---|---|
| `domain/` | Core business rules, data models, and port interfaces. The centre of the hexagon. |
| `application/` | Use-case orchestration. Coordinates domain objects and ports to fulfil a single business operation. |
| `adapters/` | Driven adapters — concrete implementations of the domain's port interfaces (Google Maps, Redis, Postgres). |
| `transport/` | Driving adapters — entry points that translate external signals (HTTP requests) into use-case calls. |
| `config/` | Environment variable loading and service configuration. |

## Dependency Rule

The only allowed import direction is **inward**:

```
transport → application → domain
adapters  → domain
```

`domain` must never import from `application`, `adapters`, or `transport`. Violations break the hexagonal architecture contract and will cause import cycles.
