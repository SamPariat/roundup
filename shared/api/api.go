// Package api holds the constants that make up the contract between Roundup's
// services: route prefixes one service calls on another, service names, and
// default ports. Values used by only one service stay in that service.
package api

// V1 is the prefix every public and internal route is mounted under.
const V1 = "/api/v1"

// UsersPath is user-service's users collection. auth-service calls POST on it
// directly at login; the gateway blocks that POST from clients.
const UsersPath = V1 + "/users"

// SquadsPath is user-service's squads collection. venue-service calls
// GET SquadsPath/{squadID} directly to check that a caller is a member.
const SquadsPath = V1 + "/squads"

// Service names, used as the logger name and in startup messages.
const (
	AuthService  = "auth-service"
	UserService  = "user-service"
	VenueService = "venue-service"
	Gateway      = "gateway"
)

// Default listen ports, used when PORT is unset. docker-compose.yml and the
// .env examples must agree with these.
const (
	AuthPort    = "3001"
	UserPort    = "3002"
	VenuePort   = "3006"
	GatewayPort = "8080"
)
