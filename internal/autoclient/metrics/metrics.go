package metrics

import (
	"fmt"
	"sync"
	"time"
)

// Collector collects and aggregates metrics for trading bots
type Collector struct {
	mu sync.RWMutex

	// Trading metrics
	OrdersSent      int64
	OrdersFilled    int64
	OrdersCancelled int64
	OrdersRejected  int64

	// Financial metrics
	TotalVolume   float64
	RealizedPnL   float64
	UnrealizedPnL float64
	Fees          float64

	// Performance metrics
	AvgFillTime    time.Duration
	StrategyUptime time.Duration
	ErrorCount     int64
	LastErrorTime  time.Time
	SuccessRate    float64

	// Production metrics (for producers)
	ProductionCount     int64
	BasicProduced       int64
	PremiumProduced     int64
	IngredientsConsumed map[string]int64

	// Timestamps
	StartTime  time.Time
	LastUpdate time.Time
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		IngredientsConsumed: make(map[string]int64),
		StartTime:           time.Now(),
		LastUpdate:          time.Now(),
	}
}

// RecordOrderSent records an order being sent
func (c *Collector) RecordOrderSent() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OrdersSent++
	c.LastUpdate = time.Now()
}

// RecordOrderFilled records an order being filled
func (c *Collector) RecordOrderFilled(volume float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OrdersFilled++
	c.TotalVolume += volume
	c.LastUpdate = time.Now()
	c.updateSuccessRate()
}

// RecordOrderCancelled records an order cancellation
func (c *Collector) RecordOrderCancelled() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OrdersCancelled++
	c.LastUpdate = time.Now()
}

// RecordOrderRejected records an order rejection
func (c *Collector) RecordOrderRejected() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OrdersRejected++
	c.LastUpdate = time.Now()
	c.updateSuccessRate()
}

// RecordPnL updates PnL metrics
func (c *Collector) RecordPnL(realized, unrealized float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.RealizedPnL = realized
	c.UnrealizedPnL = unrealized
	c.LastUpdate = time.Now()
}

// RecordFees records trading fees
func (c *Collector) RecordFees(fees float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Fees += fees
	c.LastUpdate = time.Now()
}

// RecordError records an error occurrence
func (c *Collector) RecordError() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ErrorCount++
	c.LastErrorTime = time.Now()
	c.LastUpdate = time.Now()
}

// RecordProduction records production activity
func (c *Collector) RecordProduction(product string, quantity int, isPremium bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProductionCount++
	if isPremium {
		c.PremiumProduced += int64(quantity)
	} else {
		c.BasicProduced += int64(quantity)
	}
	c.LastUpdate = time.Now()
}

// RecordIngredientConsumption records ingredient usage
func (c *Collector) RecordIngredientConsumption(ingredient string, quantity int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IngredientsConsumed[ingredient] += int64(quantity)
	c.LastUpdate = time.Now()
}

// GetSnapshot returns a snapshot of current metrics
func (c *Collector) GetSnapshot() *Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ingredientsCopy := make(map[string]int64)
	for k, v := range c.IngredientsConsumed {
		ingredientsCopy[k] = v
	}

	return &Snapshot{
		OrdersSent:          c.OrdersSent,
		OrdersFilled:        c.OrdersFilled,
		OrdersCancelled:     c.OrdersCancelled,
		OrdersRejected:      c.OrdersRejected,
		TotalVolume:         c.TotalVolume,
		RealizedPnL:         c.RealizedPnL,
		UnrealizedPnL:       c.UnrealizedPnL,
		Fees:                c.Fees,
		AvgFillTime:         c.AvgFillTime,
		ErrorCount:          c.ErrorCount,
		LastErrorTime:       c.LastErrorTime,
		SuccessRate:         c.SuccessRate,
		ProductionCount:     c.ProductionCount,
		BasicProduced:       c.BasicProduced,
		PremiumProduced:     c.PremiumProduced,
		IngredientsConsumed: ingredientsCopy,
		Uptime:              time.Since(c.StartTime),
		LastUpdate:          c.LastUpdate,
	}
}

// Reset resets all metrics
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.OrdersSent = 0
	c.OrdersFilled = 0
	c.OrdersCancelled = 0
	c.OrdersRejected = 0
	c.TotalVolume = 0
	c.RealizedPnL = 0
	c.UnrealizedPnL = 0
	c.Fees = 0
	c.ErrorCount = 0
	c.LastErrorTime = time.Time{}
	c.SuccessRate = 0
	c.ProductionCount = 0
	c.BasicProduced = 0
	c.PremiumProduced = 0
	c.IngredientsConsumed = make(map[string]int64)
	c.StartTime = time.Now()
	c.LastUpdate = time.Now()
}

// updateSuccessRate calculates the success rate (must be called with lock held)
func (c *Collector) updateSuccessRate() {
	total := c.OrdersSent
	if total == 0 {
		c.SuccessRate = 0
		return
	}
	c.SuccessRate = float64(c.OrdersFilled) / float64(total) * 100
}

// Snapshot represents a point-in-time view of metrics
type Snapshot struct {
	OrdersSent          int64
	OrdersFilled        int64
	OrdersCancelled     int64
	OrdersRejected      int64
	TotalVolume         float64
	RealizedPnL         float64
	UnrealizedPnL       float64
	Fees                float64
	AvgFillTime         time.Duration
	ErrorCount          int64
	LastErrorTime       time.Time
	SuccessRate         float64
	ProductionCount     int64
	BasicProduced       int64
	PremiumProduced     int64
	IngredientsConsumed map[string]int64
	Uptime              time.Duration
	LastUpdate          time.Time
}

// FormatSummary returns a formatted string summary of metrics
func (s *Snapshot) FormatSummary() string {
	totalProfit := s.RealizedPnL + s.UnrealizedPnL - s.Fees
	return fmt.Sprintf(`
Trading Metrics:
  Orders Sent: %d | Filled: %d | Cancelled: %d | Rejected: %d
  Success Rate: %.2f%% | Total Volume: %.2f
  
Financial:
  Realized P&L: %.2f | Unrealized P&L: %.2f
  Fees: %.2f | Net Profit: %.2f
  
Production:
  Total: %d | Basic: %d | Premium: %d
  
Performance:
  Uptime: %s | Errors: %d
  Last Update: %s
`,
		s.OrdersSent, s.OrdersFilled, s.OrdersCancelled, s.OrdersRejected,
		s.SuccessRate, s.TotalVolume,
		s.RealizedPnL, s.UnrealizedPnL,
		s.Fees, totalProfit,
		s.ProductionCount, s.BasicProduced, s.PremiumProduced,
		s.Uptime.Round(time.Second),
		s.ErrorCount,
		s.LastUpdate.Format("15:04:05"),
	)
}
