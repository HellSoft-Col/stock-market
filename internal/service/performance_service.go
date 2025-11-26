package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/HellSoft-Col/stock-market/internal/domain"
	"github.com/rs/zerolog/log"
)

type PerformanceService struct {
	teamRepo    domain.TeamRepository
	fillRepo    domain.FillRepository
	marketRepo  domain.MarketStateRepository
	broadcaster domain.Broadcaster
}

func NewPerformanceService(
	teamRepo domain.TeamRepository,
	fillRepo domain.FillRepository,
	marketRepo domain.MarketStateRepository,
	broadcaster domain.Broadcaster,
) *PerformanceService {
	return &PerformanceService{
		teamRepo:    teamRepo,
		fillRepo:    fillRepo,
		marketRepo:  marketRepo,
		broadcaster: broadcaster,
	}
}

func (s *PerformanceService) GenerateTeamReport(
	ctx context.Context,
	teamName string,
	since time.Time,
) (*domain.PerformanceReportMessage, error) {
	// Get team data
	team, err := s.teamRepo.GetByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Get fills for the team since the start time
	fills, err := s.fillRepo.GetByTeamSince(ctx, teamName, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get fills: %w", err)
	}

	// Calculate performance metrics
	var totalVolume float64
	var buyTrades, sellTrades int
	productCounts := make(map[string]int)

	for _, fill := range fills {
		volume := fill.Price * float64(fill.Quantity)
		totalVolume += volume

		// Count by product
		productCounts[fill.Product]++

		// Count by side
		if fill.Buyer == teamName {
			buyTrades++
		}
		if fill.Seller == teamName {
			sellTrades++
		}
	}

	totalTrades := len(fills)
	avgTradeSize := float64(0)
	if totalTrades > 0 {
		avgTradeSize = totalVolume / float64(totalTrades)
	}

	// Calculate inventory value at current market prices
	inventoryValue := s.calculateInventoryValue(ctx, team.Inventory)

	// Calculate total portfolio value (cash + inventory)
	finalPortfolioValue := team.CurrentBalance + inventoryValue

	// For P&L, we need to compare total portfolio value vs initial balance
	// Assuming initial inventory value was 0 (or we could track initial inventory value)
	profitLoss := finalPortfolioValue - team.InitialBalance

	roi := float64(0)
	if team.InitialBalance > 0 {
		roi = (profitLoss / team.InitialBalance) * 100
	}

	report := &domain.PerformanceReportMessage{
		Type:           "PERFORMANCE_REPORT",
		TeamName:       teamName,
		StartBalance:   team.InitialBalance,
		FinalBalance:   finalPortfolioValue, // Include inventory value
		ProfitLoss:     profitLoss,
		ROI:            roi,
		TotalTrades:    totalTrades,
		TotalVolume:    totalVolume,
		AvgTradeSize:   avgTradeSize,
		BuyTrades:      buyTrades,
		SellTrades:     sellTrades,
		Products:       productCounts,
		FinalInventory: team.Inventory,
		ServerTime:     time.Now().Format(time.RFC3339),
	}

	return report, nil
}

func (s *PerformanceService) GenerateGlobalReport(
	ctx context.Context,
	since time.Time,
) (*domain.GlobalPerformanceReportMessage, error) {
	// Get all teams
	teams, err := s.teamRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	// Get all fills since start time
	allFills, err := s.fillRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fills: %w", err)
	}

	// Filter fills by time
	var relevantFills []*domain.Fill
	for _, fill := range allFills {
		if fill.ExecutedAt.After(since) {
			relevantFills = append(relevantFills, fill)
		}
	}

	// Calculate global metrics
	var totalVolume float64
	productStats := make(map[string]*domain.ProductStats)

	for _, fill := range relevantFills {
		volume := fill.Price * float64(fill.Quantity)
		totalVolume += volume

		// Update product stats
		if productStats[fill.Product] == nil {
			productStats[fill.Product] = &domain.ProductStats{
				Product:   fill.Product,
				MinPrice:  fill.Price,
				MaxPrice:  fill.Price,
				LastPrice: fill.Price,
			}
		}

		stats := productStats[fill.Product]
		stats.TotalTrades++
		stats.TotalVolume += volume
		if fill.Price < stats.MinPrice {
			stats.MinPrice = fill.Price
		}
		if fill.Price > stats.MaxPrice {
			stats.MaxPrice = fill.Price
		}
		stats.LastPrice = fill.Price
		stats.AvgPrice = stats.TotalVolume / float64(stats.TotalTrades)
	}

	// Generate individual team reports and rank them
	var teamReports []*domain.PerformanceReportMessage
	for _, team := range teams {
		report, err := s.GenerateTeamReport(ctx, team.TeamName, since)
		if err != nil {
			log.Warn().
				Str("teamName", team.TeamName).
				Err(err).
				Msg("Failed to generate team report")
			continue
		}
		teamReports = append(teamReports, report)
	}

	// Sort teams by ROI descending
	sort.Slice(teamReports, func(i, j int) bool {
		return teamReports[i].ROI > teamReports[j].ROI
	})

	// Set ranks
	for i, report := range teamReports {
		report.Rank = i + 1
		report.TotalTeams = len(teamReports)
	}

	// Take top 10 for the report
	topTraders := teamReports
	if len(topTraders) > 10 {
		topTraders = topTraders[:10]
	}

	duration := time.Since(since).Round(time.Minute).String()

	// Convert map to format expected by message
	productStatsMap := make(map[string]domain.ProductStats)
	for product, stats := range productStats {
		productStatsMap[product] = *stats
	}

	report := &domain.GlobalPerformanceReportMessage{
		Type:         "GLOBAL_PERFORMANCE_REPORT",
		Duration:     duration,
		TotalTrades:  len(relevantFills),
		TotalVolume:  totalVolume,
		TopTraders:   topTraders,
		ProductStats: productStatsMap,
		ServerTime:   time.Now().Format(time.RFC3339),
	}

	return report, nil
}

func (s *PerformanceService) BroadcastGlobalReport(ctx context.Context, since time.Time) error {
	report, err := s.GenerateGlobalReport(ctx, since)
	if err != nil {
		return fmt.Errorf("failed to generate global report: %w", err)
	}

	// Broadcast to all connected clients
	if err := s.broadcaster.BroadcastToAll(report); err != nil {
		log.Warn().Err(err).Msg("Failed to broadcast global performance report")
	}

	log.Info().
		Int("totalTrades", report.TotalTrades).
		Float64("totalVolume", report.TotalVolume).
		Int("topTraders", len(report.TopTraders)).
		Msg("Global performance report generated and broadcasted")

	return nil
}

func (s *PerformanceService) SendTeamReport(ctx context.Context, teamName string, since time.Time) error {
	report, err := s.GenerateTeamReport(ctx, teamName, since)
	if err != nil {
		return fmt.Errorf("failed to generate team report: %w", err)
	}

	// Send to specific team
	if err := s.broadcaster.SendToClient(teamName, report); err != nil {
		return fmt.Errorf("failed to send team report: %w", err)
	}

	log.Info().
		Str("teamName", teamName).
		Float64("roi", report.ROI).
		Int("totalTrades", report.TotalTrades).
		Msg("Team performance report sent")

	return nil
}

// calculateInventoryValue computes the total value of inventory at current market prices
func (s *PerformanceService) calculateInventoryValue(ctx context.Context, inventory map[string]int) float64 {
	if len(inventory) == 0 {
		return 0
	}

	totalValue := float64(0)
	for product, quantity := range inventory {
		if quantity <= 0 {
			continue
		}

		// Get current market state for this product
		marketState, err := s.marketRepo.GetByProduct(ctx, product)
		if err != nil {
			log.Debug().
				Err(err).
				Str("product", product).
				Msg("Failed to get market state for inventory valuation, using default price")
			// Use a default price if market state not available
			totalValue += float64(quantity) * 10.0
			continue
		}

		// Use mid price if available, otherwise use best bid, or fallback to default
		var price float64
		if marketState.Mid != nil {
			price = *marketState.Mid
		} else if marketState.BestBid != nil {
			price = *marketState.BestBid
		} else {
			price = 10.0 // Default price
		}

		totalValue += float64(quantity) * price

		log.Debug().
			Str("product", product).
			Int("quantity", quantity).
			Float64("price", price).
			Float64("value", float64(quantity)*price).
			Msg("Inventory item valued")
	}

	return totalValue
}

var _ domain.PerformanceService = (*PerformanceService)(nil)
