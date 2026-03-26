# transport

Driving adapters — the left side of the hexagon. These packages translate external signals into use-case calls. They are the only entry points into the application layer.

## Sub-packages

| Package | Trigger | Calls into |
|---|---|---|
| `http/` | Inbound HTTP requests via Fiber | `VenueUseCase`, `FavouriteUseCase`, `HistoryUseCase` |

## Responsibilities

Transport packages are responsible for:
- Parsing and validating input from the external format (HTTP query params, JSON body)
- Calling the appropriate use-case method with domain types
- Translating domain errors into the appropriate external response (HTTP status codes)
- Serialising domain results back into the external format (JSON response body)

## What does NOT belong here

- Business logic — if you find yourself writing conditional domain rules in a handler, move them to the use-case
- Database queries or cache lookups — those belong in `adapters/`
- Domain model definitions — those belong in `domain/`
