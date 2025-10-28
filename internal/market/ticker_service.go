package market

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/config"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type TickerService struct {
	config      *config.Config
	marketRepo  domain.MarketStateRepository
	orderBook   domain.OrderBookRepository
	broadcaster domain.Broadcaster
	ticker      *time.Ticker
	shutdown    chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
}

func NewTickerService(
	cfg *config.Config,
	marketRepo domain.MarketStateRepository,
	orderBook domain.OrderBookRepository,
	broadcaster domain.Broadcaster,
) *TickerService {
	return &TickerService{
		config:      cfg,
		marketRepo:  marketRepo,
		orderBook:   orderBook,
		broadcaster: broadcaster,
		shutdown:    make(chan struct{}),
	}
}

func (t *TickerService) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return nil
	}

	log.Info().
		Dur("interval", t.config.Market.TickerInterval).
		Msg("Starting ticker service")

	t.running = true
	t.ticker = time.NewTicker(t.config.Market.TickerInterval)

	t.wg.Add(1)
	go t.run()

	return nil
}

func (t *TickerService) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return nil
	}

	log.Info().Msg("Stopping ticker service")

	close(t.shutdown)
	t.running = false

	if t.ticker != nil {
		t.ticker.Stop()
	}

	// Wait for goroutine to finish
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Ticker service stopped")
	case <-time.After(5 * time.Second):
		log.Warn().Msg("Ticker service stop timeout")
	}

	return nil
}

func (t *TickerService) run() {
	defer t.wg.Done()

	log.Info().Msg("Ticker service processing loop started")

	for {
		select {
		case <-t.shutdown:
			log.Info().Msg("Ticker service processing loop stopping")
			return

		case <-t.ticker.C:
			t.broadcastTickers()
		}
	}
}

func (t *TickerService) broadcastTickers() {
	products := []string{"GUACA", "SEBO", "PALTA-OIL", "FOSFO", "NUCREM", "CASCAR-ALLOY", "GTRON", "H-GUACA", "PITA"}
	serverTime := time.Now().Format(time.RFC3339)

	for _, product := range products {
		ticker := t.generateTicker(product, serverTime)
		if ticker != nil {
			if err := t.broadcaster.BroadcastToAll(ticker); err != nil {
				log.Warn().
					Str("product", product).
					Err(err).
					Msg("Failed to broadcast ticker")
			}
		}
	}

	log.Debug().
		Int("products", len(products)).
		Msg("Tickers broadcasted")
}

func (t *TickerService) generateTicker(product, serverTime string) *domain.TickerMessage {
	// Get best bid and ask from order book
	bestBidOrder := t.orderBook.GetBestBid(product)
	bestAskOrder := t.orderBook.GetBestAsk(product)

	var bestBid, bestAsk, mid *float64

	if bestBidOrder != nil && bestBidOrder.Price != nil {
		bestBid = bestBidOrder.Price
	}

	if bestAskOrder != nil && bestAskOrder.Price != nil {
		bestAsk = bestAskOrder.Price
	}

	// Calculate mid price
	if bestBid != nil && bestAsk != nil {
		midValue := (*bestBid + *bestAsk) / 2
		mid = &midValue
	}

	// Get market state for volume
	marketState, err := t.marketRepo.GetByProduct(context.Background(), product)
	if err != nil {
		log.Debug().
			Str("product", product).
			Err(err).
			Msg("Failed to get market state for ticker")
		// Continue with empty market state
		marketState = &domain.MarketState{Product: product, Volume24h: 0}
	}

	// Update market state with current best prices
	if bestBid != nil || bestAsk != nil {
		if err := t.marketRepo.UpdateBestPrices(context.Background(), product, bestBid, bestAsk); err != nil {
			log.Warn().
				Str("product", product).
				Err(err).
				Msg("Failed to update market state")
		}
	}

	ticker := &domain.TickerMessage{
		Type:       "TICKER",
		Product:    product,
		BestBid:    bestBid,
		BestAsk:    bestAsk,
		Mid:        mid,
		Volume24h:  marketState.Volume24h,
		ServerTime: serverTime,
	}

	return ticker
}
