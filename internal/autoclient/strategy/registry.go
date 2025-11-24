package strategy

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// StrategyFactory creates strategy instances
type StrategyFactory struct {
	creators map[string]StrategyCreator
}

// StrategyCreator is a function that creates a strategy instance
type StrategyCreator func(name string) Strategy

// NewStrategyFactory creates a new strategy factory
func NewStrategyFactory() *StrategyFactory {
	factory := &StrategyFactory{
		creators: make(map[string]StrategyCreator),
	}

	// Register all available strategies
	factory.Register("auto_producer", func(name string) Strategy {
		return NewAutoProducerStrategy(name)
	})

	factory.Register("market_maker", func(name string) Strategy {
		return NewMarketMakerStrategy(name)
	})

	factory.Register("random_trader", func(name string) Strategy {
		return NewRandomTraderStrategy(name)
	})

	factory.Register("liquidity_provider", func(name string) Strategy {
		return NewLiquidityProviderStrategy(name)
	})

	factory.Register("momentum_trader", func(name string) Strategy {
		return NewMomentumStrategy(name)
	})

	factory.Register("arbitrage", func(name string) Strategy {
		return NewArbitrageStrategy(name)
	})

	factory.Register("deepseek", func(name string) Strategy {
		return NewDeepSeekStrategy(name)
	})

	factory.Register("hybrid", func(name string) Strategy {
		return NewHybridStrategy(name)
	})

	factory.Register("buffett", func(name string) Strategy {
		return NewBuffettStrategy(name)
	})

	log.Info().
		Int("strategies", len(factory.creators)).
		Msg("Strategy factory initialized")

	return factory
}

// Register registers a new strategy creator
func (f *StrategyFactory) Register(strategyType string, creator StrategyCreator) {
	f.creators[strategyType] = creator
}

// Create creates a strategy instance
func (f *StrategyFactory) Create(strategyType string, name string) (Strategy, error) {
	creator, exists := f.creators[strategyType]
	if !exists {
		return nil, fmt.Errorf("unknown strategy type: %s", strategyType)
	}

	strategy := creator(name)

	log.Info().
		Str("type", strategyType).
		Str("name", name).
		Msg("Strategy created")

	return strategy, nil
}

// ListAvailable returns all available strategy types
func (f *StrategyFactory) ListAvailable() []string {
	types := make([]string, 0, len(f.creators))
	for strategyType := range f.creators {
		types = append(types, strategyType)
	}
	return types
}
