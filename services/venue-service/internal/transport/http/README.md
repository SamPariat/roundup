# transport/http

Fiber HTTP server setup and all HTTP handler methods.

## Files

| File | Purpose |
|---|---|
| `server.go` | Builds and returns the configured `*fiber.App`. Registers all routes, attaches middleware, and sets up the global error handler. |
| `venue_handler.go` | Handles `GET /venues/search` and `GET /venues/:placeId`. Calls `VenueUseCase`. |
| `favourite_handler.go` | Handles `POST /venues/favourites`, `DELETE /venues/favourites`, `GET /venues/favourites`. Calls `FavouriteUseCase`. |
| `history_handler.go` | Handles `GET /venues/history`. Calls `HistoryUseCase`. |

## Routes

| Method | Path | Handler |
|---|---|---|
| GET | `/venues/search` | `VenueHandler.Search` |
| GET | `/venues/:placeId` | `VenueHandler.GetDetail` |
| POST | `/venues/favourites` | `FavouriteHandler.Add` |
| DELETE | `/venues/favourites` | `FavouriteHandler.Remove` |
| GET | `/venues/favourites` | `FavouriteHandler.List` |
| GET | `/venues/history` | `HistoryHandler.List` |
| GET | `/health` | inline — returns `{"status":"ok"}` |

## Error mapping

`mapDomainError` (in `venue_handler.go`) is the single place where domain sentinel errors are translated to HTTP status codes:

| Domain error | HTTP status |
|---|---|
| `ErrVenueNotFound` | 404 |
| `ErrInvalidPlaceID`, `ErrInvalidCoordinates` | 400 |
| `ErrAlreadyFavourited` | 409 |
| `ErrProviderUnavailable` | 502 |
| anything else | 500 |

## Sub-packages

| Package | Purpose |
|---|---|
| `dto/` | Request and response structs with `json:` tags and input validation. These types never appear in the domain or application layers. |
| `middleware/` | Fiber middleware: request ID injection and structured request logging. |
