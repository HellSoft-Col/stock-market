package domain

import "errors"

// Error codes
const (
	ErrAuthFailed             = "AUTH_FAILED"
	ErrInvalidOrder           = "INVALID_ORDER"
	ErrInvalidProduct         = "INVALID_PRODUCT"
	ErrInvalidQuantity        = "INVALID_QUANTITY"
	ErrDuplicateOrderID       = "DUPLICATE_ORDER_ID"
	ErrUnauthorizedProduction = "UNAUTHORIZED_PRODUCTION"
	ErrOfferExpired           = "OFFER_EXPIRED"
	ErrRateLimitExceeded      = "RATE_LIMIT_EXCEEDED"
	ErrServiceUnavailable     = "SERVICE_UNAVAILABLE"
	ErrInsufficientInventory  = "INSUFFICIENT_INVENTORY"
	ErrInvalidMessage         = "INVALID_MESSAGE"
)

// Domain errors
var (
	ErrTeamNotFound     = errors.New("team not found")
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidOrderSide = errors.New("invalid order side")
	ErrInvalidOrderMode = errors.New("invalid order mode")
	ErrProductNotFound  = errors.New("product not found")
)
