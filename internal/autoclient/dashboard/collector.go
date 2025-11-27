package dashboard

import (
	"sync"
	"time"
)

// MetricsCollector collects and aggregates metrics from trading agents
type MetricsCollector struct {
	dashboard *Dashboard
	mu        sync.RWMutex
	metrics   map[string]*TraderMetrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(dashboard *Dashboard) *MetricsCollector {
	return &MetricsCollector{
		dashboard: dashboard,
		metrics:   make(map[string]*TraderMetrics),
	}
}

// InitializeTrader initializes metrics for a new trader
func (c *MetricsCollector) InitializeTrader(teamName, strategy string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[teamName] = &TraderMetrics{
		TeamName:       teamName,
		Strategy:       strategy,
		Status:         "active",
		LastActionTime: time.Now(),
	}

	c.syncToDashboard(teamName)
}

// UpdateBalance updates the balance for a trader
func (c *MetricsCollector) UpdateBalance(teamName string, balance float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.Balance = balance
		c.syncToDashboard(teamName)
	}
}

// UpdateInventoryValue updates the inventory value for a trader
func (c *MetricsCollector) UpdateInventoryValue(teamName string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.InventoryValue = value
		c.syncToDashboard(teamName)
	}
}

// UpdateNetWorth updates the net worth for a trader
func (c *MetricsCollector) UpdateNetWorth(teamName string, netWorth float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.NetWorth = netWorth
		c.syncToDashboard(teamName)
	}
}

// UpdatePnL updates the P&L for a trader
func (c *MetricsCollector) UpdatePnL(teamName string, pnl, pnlPercent float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.PnL = pnl
		m.PnLPercent = pnlPercent
		c.syncToDashboard(teamName)
	}
}

// IncrementOrders increments the order count for a trader
func (c *MetricsCollector) IncrementOrders(teamName string, count int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.OrdersPlaced += count
		c.syncToDashboard(teamName)
	}
}

// IncrementFills increments the fill count for a trader
func (c *MetricsCollector) IncrementFills(teamName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.FillsReceived++
		c.syncToDashboard(teamName)
	}
}

// UpdateActiveOrders updates the active order count for a trader
func (c *MetricsCollector) UpdateActiveOrders(teamName string, count int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.ActiveOrders = count
		c.syncToDashboard(teamName)
	}
}

// RecordAction records a trading action
func (c *MetricsCollector) RecordAction(teamName, action string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.LastAction = action
		m.LastActionTime = time.Now()
		c.syncToDashboard(teamName)
	}
}

// IncrementAIDecisions increments the AI decision count
func (c *MetricsCollector) IncrementAIDecisions(teamName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.AIDecisions++
		c.syncToDashboard(teamName)
	}
}

// IncrementProduction increments the production count
func (c *MetricsCollector) IncrementProduction(teamName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.ProductionCount++
		c.syncToDashboard(teamName)
	}
}

// IncrementErrors increments the error count and updates status
func (c *MetricsCollector) IncrementErrors(teamName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.ErrorCount++
		if m.ErrorCount > 5 {
			m.Status = "error"
		} else if m.ErrorCount > 2 {
			m.Status = "degraded"
		}
		c.syncToDashboard(teamName)
	}
}

// ResetErrors resets the error count and status
func (c *MetricsCollector) ResetErrors(teamName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.ErrorCount = 0
		m.Status = "active"
		c.syncToDashboard(teamName)
	}
}

// UpdateStatus updates the status for a trader
func (c *MetricsCollector) UpdateStatus(teamName, status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.metrics[teamName]; ok {
		m.Status = status
		c.syncToDashboard(teamName)
	}
}

// UpdateAll updates all metrics at once
func (c *MetricsCollector) UpdateAll(teamName string, metrics *TraderMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics[teamName] = metrics
	c.syncToDashboard(teamName)
}

// syncToDashboard syncs metrics to the dashboard (must be called with lock held)
func (c *MetricsCollector) syncToDashboard(teamName string) {
	if c.dashboard != nil {
		if m, ok := c.metrics[teamName]; ok {
			// Create a copy to avoid race conditions
			metricsCopy := *m
			c.dashboard.UpdateTrader(teamName, &metricsCopy)
		}
	}
}

// GetMetrics returns a copy of all metrics
func (c *MetricsCollector) GetMetrics(teamName string) *TraderMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.metrics[teamName]; ok {
		metricsCopy := *m
		return &metricsCopy
	}
	return nil
}
