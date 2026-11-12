// Package api holds the contract between Roundup's services: route prefixes
// one service calls on another, the bodies those calls and shared responses
// carry, service names, and default ports. Values used by only one service
// stay in that service.
package api

// V1 is the prefix every public and internal route is mounted under.
const V1 = "/api/v1"

// InternalV1 is the prefix of routes only other services call, such as
// scheduler-service's sweeps. The gateway never routes it.
const InternalV1 = "/internal/v1"

// UsersPath is user-service's users collection. auth-service calls POST on it
// directly at login; the gateway blocks that POST from clients.
const UsersPath = V1 + "/users"

// SquadsPath is user-service's squads collection. venue-service calls
// GET SquadsPath/{squadID} directly to check that a caller is a member.
const SquadsPath = V1 + "/squads"

// VenuesPath is venue-service's venues collection. event-service calls
// GET VenuesPath/{placeID} directly to snapshot a nominated place.
const VenuesPath = V1 + "/venues"

// ClientHeader is the request header a client names itself in: ios, android
// or web. The gateway lets browsers send it, and auth-service reads it to keep
// a web client's refresh token in an httpOnly cookie.
const ClientHeader = "X-Client"

// ClientWeb is the ClientHeader value of the browser app.
const ClientWeb = "web"

// Service names, used as the logger name and in startup messages.
const (
	AuthService  = "auth-service"
	UserService  = "user-service"
	VenueService = "venue-service"
	EventService = "event-service"
	Gateway      = "gateway"
)

// Default listen ports, used when PORT is unset. docker-compose.yml and the
// .env examples must agree with these.
const (
	AuthPort    = "3001"
	UserPort    = "3002"
	EventPort   = "3003"
	VenuePort   = "3006"
	GatewayPort = "8080"
)
