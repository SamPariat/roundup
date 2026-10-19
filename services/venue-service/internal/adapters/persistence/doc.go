// Package persistence implements the venue-service repository ports on top of the
// sqlc-generated queries in ./db. Driver errors such as unique violations and
// missing rows are mapped to domain sentinels here, so nothing above this
// package sees a pgx type.
package persistence
