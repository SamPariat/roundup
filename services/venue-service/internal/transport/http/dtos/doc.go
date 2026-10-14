// Package dtos contains the request and response types for the venue-service HTTP API.
// All types in this package exist at the HTTP boundary — domain types are mapped into
// these structs in handlers and never returned directly to clients. Response DTOs are
// built only by the functions in mappers.go.
package dtos
