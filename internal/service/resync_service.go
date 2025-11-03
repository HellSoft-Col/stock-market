package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/HellSoft-Col/stock-market/internal/domain"
)

type ResyncService struct {
	fillRepo domain.FillRepository
}

func NewResyncService(fillRepo domain.FillRepository) *ResyncService {
	return &ResyncService{
		fillRepo: fillRepo,
	}
}

func (s *ResyncService) GenerateEventDelta(ctx context.Context, teamName string, since time.Time) (*domain.EventDeltaMessage, error) {
	// Get all fills for this team since the specified time
	fills, err := s.fillRepo.GetByTeamSince(ctx, teamName, since)
	if err != nil {
		log.Error().
			Str("teamName", teamName).
			Time("since", since).
			Err(err).
			Msg("Failed to get fills for resync")
		return nil, err
	}

	// Convert fills to FILL messages
	events := make([]domain.FillMessage, 0, len(fills))
	for _, fill := range fills {
		// Determine which side this team was on
		var side string
		var counterparty string
		var counterpartyMessage string
		var clOrdID string

		if fill.Buyer == teamName {
			side = "BUY"
			counterparty = fill.Seller
			counterpartyMessage = fill.SellerMessage
			clOrdID = fill.BuyerClOrdID
		} else {
			side = "SELL"
			counterparty = fill.Buyer
			counterpartyMessage = fill.BuyerMessage
			clOrdID = fill.SellerClOrdID
		}

		fillMsg := domain.FillMessage{
			Type:                "FILL",
			ClOrdID:             clOrdID,
			FillQty:             fill.Quantity,
			FillPrice:           fill.Price,
			Side:                side,
			Product:             fill.Product,
			Counterparty:        counterparty,
			CounterpartyMessage: counterpartyMessage,
			ServerTime:          fill.ExecutedAt.Format(time.RFC3339),
		}

		events = append(events, fillMsg)
	}

	eventDelta := &domain.EventDeltaMessage{
		Type:       "EVENT_DELTA",
		Events:     events,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("teamName", teamName).
		Time("since", since).
		Int("fillCount", len(fills)).
		Msg("Generated event delta for resync")

	return eventDelta, nil
}

var _ domain.ResyncService = (*ResyncService)(nil)
