package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKey              string             `bson:"apiKey" json:"apiKey"`
	TeamName            string             `bson:"teamName" json:"teamName"`
	Species             string             `bson:"species" json:"species"`
	InitialBalance      float64            `bson:"initialBalance" json:"initialBalance"`
	CurrentBalance      float64            `bson:"currentBalance" json:"currentBalance"`
	AuthorizedProducts  []string           `bson:"authorizedProducts" json:"authorizedProducts"`
	Inventory           map[string]int     `bson:"inventory" json:"inventory"` // Product -> Quantity
	Recipes             map[string]Recipe  `bson:"recipes" json:"recipes"`
	Role                TeamRole           `bson:"role" json:"role"`
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	LastLogin           time.Time          `bson:"lastLogin" json:"lastLogin"`
	LastInventoryUpdate time.Time          `bson:"lastInventoryUpdate" json:"lastInventoryUpdate"`
}

type Recipe struct {
	Type         string         `bson:"type" json:"type"`
	Ingredients  map[string]int `bson:"ingredients" json:"ingredients"`
	PremiumBonus float64        `bson:"premiumBonus" json:"premiumBonus"`
}

type TeamRole struct {
	Branches    int     `bson:"branches" json:"branches"`
	MaxDepth    int     `bson:"maxDepth" json:"maxDepth"`
	Decay       float64 `bson:"decay" json:"decay"`
	Budget      float64 `bson:"budget" json:"budget"`
	BaseEnergy  float64 `bson:"baseEnergy" json:"baseEnergy"`
	LevelEnergy float64 `bson:"levelEnergy" json:"levelEnergy"`
}

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ClOrdID    string             `bson:"clOrdID" json:"clOrdID"`
	TeamName   string             `bson:"teamName" json:"teamName"`
	Side       string             `bson:"side" json:"side"` // "BUY" or "SELL"
	Mode       string             `bson:"mode" json:"mode"` // "MARKET" or "LIMIT"
	Product    string             `bson:"product" json:"product"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	Price      *float64           `bson:"price,omitempty" json:"price,omitempty"` // For LIMIT orders
	LimitPrice *float64           `bson:"limitPrice,omitempty" json:"limitPrice,omitempty"`
	Message    string             `bson:"message" json:"message"`
	DebugMode  string             `bson:"debugMode,omitempty" json:"debugMode,omitempty"` // "AUTO_ACCEPT", "TEAM_ONLY"
	Status     string             `bson:"status" json:"status"`                           // "PENDING", "FILLED", "PARTIALLY_FILLED", "CANCELLED"
	FilledBy   *string            `bson:"filledBy,omitempty" json:"filledBy,omitempty"`
	FilledQty  int                `bson:"filledQty" json:"filledQty"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	FilledAt   *time.Time         `bson:"filledAt,omitempty" json:"filledAt,omitempty"`
	ExpiresAt  *time.Time         `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
}

type Fill struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FillID          string             `bson:"fillID" json:"fillID"`
	BuyerClOrdID    string             `bson:"buyerClOrdID" json:"buyerClOrdID"`
	SellerClOrdID   string             `bson:"sellerClOrdID" json:"sellerClOrdID"`
	Buyer           string             `bson:"buyer" json:"buyer"`
	Seller          string             `bson:"seller" json:"seller"`
	Product         string             `bson:"product" json:"product"`
	Quantity        int                `bson:"quantity" json:"quantity"`
	Price           float64            `bson:"price" json:"price"`
	BuyerMessage    string             `bson:"buyerMessage" json:"buyerMessage"`
	SellerMessage   string             `bson:"sellerMessage" json:"sellerMessage"`
	ExecutedAt      time.Time          `bson:"executedAt" json:"executedAt"`
	TournamentPhase string             `bson:"tournamentPhase,omitempty" json:"tournamentPhase,omitempty"`
}

type MarketState struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Product        string             `bson:"product" json:"product"`
	BestBid        *float64           `bson:"bestBid,omitempty" json:"bestBid,omitempty"`
	BestAsk        *float64           `bson:"bestAsk,omitempty" json:"bestAsk,omitempty"`
	Mid            *float64           `bson:"mid,omitempty" json:"mid,omitempty"`
	LastTradePrice *float64           `bson:"lastTradePrice,omitempty" json:"lastTradePrice,omitempty"`
	Volume24h      int                `bson:"volume24h" json:"volume24h"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Product struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	BasePrice  float64 `json:"basePrice"`
	Volatility float64 `json:"volatility"`
}

type InventoryTransaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamName  string             `bson:"teamName" json:"teamName"`
	Product   string             `bson:"product" json:"product"`
	Change    int                `bson:"change" json:"change"` // Positive for production/buy, negative for sell
	Reason    string             `bson:"reason" json:"reason"` // "PRODUCTION", "TRADE_BUY", "TRADE_SELL", "INITIAL"
	OrderID   string             `bson:"orderID,omitempty" json:"orderID,omitempty"`
	FillID    string             `bson:"fillID,omitempty" json:"fillID,omitempty"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
