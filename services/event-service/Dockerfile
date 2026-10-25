# Build context must be the monorepo root (not services/event-service), since this
# module depends on ../../shared via a replace directive:
#   docker build -f services/event-service/Dockerfile -t roundup-event-service .
#
# The image holds no env files and no APP_ENV: set APP_ENV (beta|prod) and the
# service's variables at runtime, or it refuses to start. The same image is
# promoted from beta to prod unchanged.
#
# Neither stage runs as root: the build stage uses an unprivileged "builder"
# user, and the runtime image is distroless with its built-in "nonroot" user.

FROM golang:1.27-alpine AS build
RUN adduser -D -u 10001 builder
USER builder
WORKDIR /home/builder/src

COPY --chown=builder:builder shared ./shared
COPY --chown=builder:builder services/event-service ./services/event-service
WORKDIR /home/builder/src/services/event-service

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /home/builder/event-service ./cmd

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /home/builder/event-service /event-service

USER nonroot:nonroot
EXPOSE 3003
ENTRYPOINT ["/event-service"]
