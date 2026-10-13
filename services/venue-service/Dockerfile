# Build context must be the monorepo root (not services/venue-service), since this
# module depends on ../../shared via a replace directive:
#   docker build -f services/venue-service/Dockerfile -t roundup-venue-service .
#
# Neither stage runs as root: the build stage uses an unprivileged "builder"
# user, and the runtime image is distroless with its built-in "nonroot" user.

FROM golang:1.26-alpine AS build
RUN adduser -D -u 10001 builder
USER builder
WORKDIR /home/builder/src

COPY --chown=builder:builder shared ./shared
COPY --chown=builder:builder services/venue-service ./services/venue-service
WORKDIR /home/builder/src/services/venue-service

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /home/builder/venue-service ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /home/builder/venue-service /venue-service

USER nonroot:nonroot
EXPOSE 3006
ENTRYPOINT ["/venue-service"]
