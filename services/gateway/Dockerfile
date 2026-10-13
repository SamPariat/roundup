# Build context must be the monorepo root (not services/gateway), since this
# module depends on ../../shared via a replace directive:
#   docker build -f services/gateway/Dockerfile -t roundup-gateway .
#
# Neither stage runs as root: the build stage uses an unprivileged "builder"
# user, and the runtime image is distroless with its built-in "nonroot" user.

FROM golang:1.26-alpine AS build
RUN adduser -D -u 10001 builder
USER builder
WORKDIR /home/builder/src

COPY --chown=builder:builder shared ./shared
COPY --chown=builder:builder services/gateway ./services/gateway
WORKDIR /home/builder/src/services/gateway

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /home/builder/gateway ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /home/builder/gateway /gateway

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/gateway"]
