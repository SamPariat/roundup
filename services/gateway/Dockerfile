# Build context must be the monorepo root (not services/gateway), since this
# module depends on ../../shared/go via a replace directive:
#   docker build -f services/gateway/Dockerfile -t roundup-gateway .

FROM golang:1.26-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src

COPY shared/go ./shared/go
COPY services/gateway ./services/gateway
WORKDIR /src/services/gateway

RUN go build -trimpath -ldflags="-s -w" -o /gateway ./cmd

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /gateway /gateway

EXPOSE 8080
ENTRYPOINT ["/gateway"]
