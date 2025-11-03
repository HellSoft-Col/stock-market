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
	DebugMode  string   `json:"debugMode,omitempty"` // "AUTO_ACCEPT", "TEAM_ONLY", ""
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

type CancelMessage struct {
	Type    string `json:"type"`
	ClOrdID string `json:"clOrdID"`
}

type AdminCancelAllOrdersMessage struct {
	Type string `json:"type"`
}

type AdminBroadcastMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type AdminCreateOrderMessage struct {
	Type       string   `json:"type"`
	TeamName   string   `json:"teamName"`
	Side       string   `json:"side"`
	Mode       string   `json:"mode"`
	Product    string   `json:"product"`
	Qty        int      `json:"qty"`
	LimitPrice *float64 `json:"limitPrice,omitempty"`
	Message    string   `json:"message,omitempty"`
}

type ExportDataMessage struct {
	Type      string `json:"type"`
	DataType  string `json:"dataType"` // "orders", "fills", "teams", "all"
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type GetAvailableTeamsMessage struct {
	Type string `json:"type"`
}

type GetTeamActivityMessage struct {
	Type     string `json:"type"`
	TeamName string `json:"teamName,omitempty"`
}

type GetAllTeamsMessage struct {
	Type string `json:"type"`
}

type UpdateTeamMessage struct {
	Type      string         `json:"type"`
	TeamName  string         `json:"teamName"`
	Balance   float64        `json:"balance"`
	Inventory map[string]int `json:"inventory"`
}

type ResetTeamBalanceMessage struct {
	Type     string `json:"type"`
	TeamName string `json:"teamName"`
}

type ResetTeamInventoryMessage struct {
	Type     string `json:"type"`
	TeamName string `json:"teamName"`
}

type ResetTeamProductionMessage struct {
	Type     string `json:"type"`
	TeamName string `json:"teamName"`
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

type BalanceUpdateMessage struct {
	Type       string  `json:"type"`
	Balance    float64 `json:"balance"`
	ServerTime string  `json:"serverTime"`
}

type OrderBookUpdateMessage struct {
	Type       string          `json:"type"`
	Product    string          `json:"product"`
	BuyOrders  []*OrderSummary `json:"buyOrders"`
	SellOrders []*OrderSummary `json:"sellOrders"`
	ServerTime string          `json:"serverTime"`
}

type OrderSummary struct {
	ClOrdID   string   `json:"clOrdID"`
	TeamName  string   `json:"teamName"`
	Side      string   `json:"side"`
	Mode      string   `json:"mode"`
	Product   string   `json:"product"`
	Quantity  int      `json:"quantity"`
	Price     *float64 `json:"price,omitempty"`
	FilledQty int      `json:"filledQty"`
	Message   string   `json:"message,omitempty"`
	CreatedAt string   `json:"createdAt"`
}

type AllOrdersMessage struct {
	Type       string          `json:"type"`
	Orders     []*OrderSummary `json:"orders"`
	ServerTime string          `json:"serverTime"`
}

type SessionInfo struct {
	TeamName      string `json:"teamName"`
	RemoteAddr    string `json:"remoteAddr"`
	UserAgent     string `json:"userAgent,omitempty"`
	ClientType    string `json:"clientType"`
	ConnectedAt   string `json:"connectedAt"`
	LastActivity  string `json:"lastActivity,omitempty"`
	Authenticated bool   `json:"authenticated"`
}

type ConnectedSessionsMessage struct {
	Type       string         `json:"type"`
	Sessions   []*SessionInfo `json:"sessions"`
	ServerTime string         `json:"serverTime"`
}

type ServerStatsMessage struct {
	Type       string                 `json:"type"`
	Stats      map[string]interface{} `json:"stats"`
	ServerTime string                 `json:"serverTime"`
}

type PerformanceReportMessage struct {
	Type           string         `json:"type"`
	TeamName       string         `json:"teamName"`
	StartBalance   float64        `json:"startBalance"`
	FinalBalance   float64        `json:"finalBalance"`
	ProfitLoss     float64        `json:"profitLoss"`
	ROI            float64        `json:"roi"` // Return on Investment %
	TotalTrades    int            `json:"totalTrades"`
	TotalVolume    float64        `json:"totalVolume"`
	AvgTradeSize   float64        `json:"avgTradeSize"`
	BuyTrades      int            `json:"buyTrades"`
	SellTrades     int            `json:"sellTrades"`
	Products       map[string]int `json:"products"` // Product -> trade count
	FinalInventory map[string]int `json:"finalInventory"`
	Rank           int            `json:"rank,omitempty"`
	TotalTeams     int            `json:"totalTeams,omitempty"`
	ServerTime     string         `json:"serverTime"`
}

type GlobalPerformanceReportMessage struct {
	Type         string                      `json:"type"`
	Duration     string                      `json:"duration"`
	TotalTrades  int                         `json:"totalTrades"`
	TotalVolume  float64                     `json:"totalVolume"`
	TopTraders   []*PerformanceReportMessage `json:"topTraders"`
	ProductStats map[string]ProductStats     `json:"productStats"`
	ServerTime   string                      `json:"serverTime"`
}

type ProductStats struct {
	Product     string  `json:"product"`
	TotalTrades int     `json:"totalTrades"`
	TotalVolume float64 `json:"totalVolume"`
	AvgPrice    float64 `json:"avgPrice"`
	MinPrice    float64 `json:"minPrice"`
	MaxPrice    float64 `json:"maxPrice"`
	LastPrice   float64 `json:"lastPrice"`
}

type BroadcastNotificationMessage struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Sender     string `json:"sender"` // "admin"
	ServerTime string `json:"serverTime"`
}

type ExportDataResponse struct {
	Type       string      `json:"type"`
	DataType   string      `json:"dataType"`
	Data       interface{} `json:"data"`
	Count      int         `json:"count"`
	ServerTime string      `json:"serverTime"`
}

type AdminActionResponse struct {
	Type       string `json:"type"`
	Action     string `json:"action"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Count      int    `json:"count,omitempty"`
	ServerTime string `json:"serverTime"`
}

type AvailableTeamsResponse struct {
	Type       string     `json:"type"`
	Teams      []TeamInfo `json:"teams"`
	Count      int        `json:"count"`
	ServerTime string     `json:"serverTime"`
}

type TeamInfo struct {
	TeamName           string   `json:"teamName"`
	Species            string   `json:"species"`
	Connected          bool     `json:"connected"`
	Balance            float64  `json:"balance"`
	ActiveOrders       int      `json:"activeOrders"`
	AuthorizedProducts []string `json:"authorizedProducts,omitempty"`
}

type TeamActivityResponse struct {
	Type       string           `json:"type"`
	TeamName   string           `json:"teamName"`
	Activities []ActivityRecord `json:"activities"`
	Count      int              `json:"count"`
	ServerTime string           `json:"serverTime"`
}

type ActivityRecord struct {
	Timestamp string  `json:"timestamp"`
	Action    string  `json:"action"` // "ORDER", "FILL", "CANCEL", "PRODUCTION", "LOGIN", "LOGOUT"
	Details   string  `json:"details"`
	Product   string  `json:"product,omitempty"`
	Quantity  int     `json:"quantity,omitempty"`
	Price     float64 `json:"price,omitempty"`
}

type AllTeamsResponse struct {
	Type       string      `json:"type"`
	Teams      []*TeamData `json:"teams"`
	Count      int         `json:"count"`
	ServerTime string      `json:"serverTime"`
}

type TeamData struct {
	TeamName           string         `json:"teamName"`
	Species            string         `json:"species"`
	InitialBalance     float64        `json:"initialBalance"`
	CurrentBalance     float64        `json:"currentBalance"`
	Inventory          map[string]int `json:"inventory"`
	AuthorizedProducts []string       `json:"authorizedProducts"`
	Connected          bool           `json:"connected"`
}

type TeamUpdatedResponse struct {
	Type       string `json:"type"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ServerTime string `json:"serverTime"`
}

// Base message for parsing
type BaseMessage struct {
	Type string `json:"type"`
}
