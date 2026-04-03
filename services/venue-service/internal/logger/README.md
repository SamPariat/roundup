# logger

Zap logger construction for the venue-service.

## Files

| File | Purpose |
|---|---|
| `logger.go` | Builds and returns a `*zap.Logger` from the log level string in `Config`. |

## Usage

```go
cfg := config.Load()
log := logger.New(cfg.LogLevel)
```

`log` is then passed into adapter and handler constructors — never accessed globally.

## Behaviour

- Uses `zap.NewProductionConfig()` — JSON output, no development-mode caller info spam.
- Panics at startup if `LOG_LEVEL` is not a valid zapcore level (`debug`, `info`, `warn`, `error`).
- Defaults to `info` when `LOG_LEVEL` is unset (handled by the `config` package).

## Log levels

| Level | When to use |
|---|---|
| `debug` | Local development only — high-volume, request-level detail |
| `info` | Default — service lifecycle events, significant state changes |
| `warn` | Recoverable anomalies (cache miss storms, slow queries) |
| `error` | Failures that affect a request but don't crash the service |
