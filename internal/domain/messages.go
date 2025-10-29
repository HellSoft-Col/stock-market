package domain

import "time"

// Client to Server Messages

type LoginMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	TZ    string `json:"tz,omitempty"`
}

type OrderMessage struct {
	Type       string   `json:"type"`
	ClOrdID    string   `json:"clOrdID"`
	Side       string   `json:"side"`
	Mode       string   `json:"mode"`
	Product    string   `json:"product"`
	Qty        int      `json:"qty"`
	LimitPrice *float64 `json:"limitPrice,omitempty"`
	ExpiresAt  *string  `json:"expiresAt,omitempty"`
	Message    string   `json:"message,omitempty"`
}

type ProductionUpdateMessage struct {
	Type     string `json:"type"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

type AcceptOfferMessage struct {
	Type            string  `json:"type"`
	OfferID         string  `json:"offerId"`
	Accept          bool    `json:"accept"`
	QuantityOffered int     `json:"quantityOffered,omitempty"`
	PriceOffered    float64 `json:"priceOffered,omitempty"`
}

type ResyncMessage struct {
	Type     string `json:"type"`
	LastSync string `json:"lastSync"`
}

// Server to Client Messages

type LoginOKMessage struct {
	Type               string            `json:"type"`
	Team               string            `json:"team"`
	Species            string            `json:"species"`
	InitialBalance     float64           `json:"initialBalance"`
	CurrentBalance     float64           `json:"currentBalance"`
	Inventory          map[string]int    `json:"inventory"`
	AuthorizedProducts []string          `json:"authorizedProducts"`
	Recipes            map[string]Recipe `json:"recipes"`
	Role               TeamRole          `json:"role"`
	ServerTime         string            `json:"serverTime"`
}

type FillMessage struct {
	Type                string  `json:"type"`
	ClOrdID             string  `json:"clOrdID"`
	FillQty             int     `json:"fillQty"`
	FillPrice           float64 `json:"fillPrice"`
	Side                string  `json:"side"`
	Product             string  `json:"product"`
	Counterparty        string  `json:"counterparty"`
	CounterpartyMessage string  `json:"counterpartyMessage"`
	ServerTime          string  `json:"serverTime"`
	RemainingQty        int     `json:"remainingQty,omitempty"` // For partial fills
	TotalQty            int     `json:"totalQty,omitempty"`     // For partial fills
}

type TickerMessage struct {
	Type       string   `json:"type"`
	Product    string   `json:"product"`
	BestBid    *float64 `json:"bestBid,omitempty"`
	BestAsk    *float64 `json:"bestAsk,omitempty"`
	Mid        *float64 `json:"mid,omitempty"`
	Volume24h  int      `json:"volume24h"`
	ServerTime string   `json:"serverTime"`
}

type OfferMessage struct {
	Type              string    `json:"type"`
	OfferID           string    `json:"offerId"`
	Buyer             string    `json:"buyer"`
	Product           string    `json:"product"`
	QuantityRequested int       `json:"quantityRequested"`
	MaxPrice          float64   `json:"maxPrice"`
	ExpiresIn         *int      `json:"expiresIn,omitempty"` // milliseconds, nil = no expiration
	Timestamp         time.Time `json:"timestamp"`
}

type ErrorMessage struct {
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	Reason    string    `json:"reason"`
	ClOrdID   string    `json:"clOrdID,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type EventDeltaMessage struct {
	Type       string        `json:"type"`
	Events     []FillMessage `json:"events"`
	ServerTime string        `json:"serverTime"`
}

type OrderAckMessage struct {
	Type       string `json:"type"`
	ClOrdID    string `json:"clOrdID"`
	Status     string `json:"status"`
	ServerTime string `json:"serverTime"`
}

type InventoryUpdateMessage struct {
	Type       string         `json:"type"`
	Inventory  map[string]int `json:"inventory"`
	ServerTime string         `json:"serverTime"`
}

// Base message for parsing
type BaseMessage struct {
	Type string `json:"type"`
}
