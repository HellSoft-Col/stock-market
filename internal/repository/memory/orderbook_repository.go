package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/yourusername/avocado-exchange-server/internal/domain"
)

type OrderBook struct {
	Product    string
	BuyOrders  []*domain.Order // Sorted by price DESC (highest first)
	SellOrders []*domain.Order // Sorted by price ASC (lowest first)
}

type OrderBookRepository struct {
	orderBooks    map[string]*OrderBook // product -> OrderBook
	recentSellers map[string][]string   // product -> []teamName
	mu            sync.RWMutex
}

func NewOrderBookRepository() *OrderBookRepository {
	return &OrderBookRepository{
		orderBooks:    make(map[string]*OrderBook),
		recentSellers: make(map[string][]string),
	}
}

func (r *OrderBookRepository) AddOrder(product, side string, order *domain.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get or create order book for product
	if r.orderBooks[product] == nil {
		r.orderBooks[product] = &OrderBook{Product: product}
	}

	book := r.orderBooks[product]

	switch side {
	case "BUY":
		book.BuyOrders = append(book.BuyOrders, order)
		// Sort by price DESC (highest bid first)
		sort.Slice(book.BuyOrders, func(i, j int) bool {
			if book.BuyOrders[i].Price == nil && book.BuyOrders[j].Price == nil {
				return book.BuyOrders[i].CreatedAt.Before(book.BuyOrders[j].CreatedAt)
			}
			if book.BuyOrders[i].Price == nil {
				return false // Market orders go to end
			}
			if book.BuyOrders[j].Price == nil {
				return true
			}
			if *book.BuyOrders[i].Price == *book.BuyOrders[j].Price {
				return book.BuyOrders[i].CreatedAt.Before(book.BuyOrders[j].CreatedAt)
			}
			return *book.BuyOrders[i].Price > *book.BuyOrders[j].Price
		})

	case "SELL":
		book.SellOrders = append(book.SellOrders, order)
		// Sort by price ASC (lowest ask first)
		sort.Slice(book.SellOrders, func(i, j int) bool {
			if book.SellOrders[i].Price == nil && book.SellOrders[j].Price == nil {
				return book.SellOrders[i].CreatedAt.Before(book.SellOrders[j].CreatedAt)
			}
			if book.SellOrders[i].Price == nil {
				return false // Market orders go to end
			}
			if book.SellOrders[j].Price == nil {
				return true
			}
			if *book.SellOrders[i].Price == *book.SellOrders[j].Price {
				return book.SellOrders[i].CreatedAt.Before(book.SellOrders[j].CreatedAt)
			}
			return *book.SellOrders[i].Price < *book.SellOrders[j].Price
		})

		// Track recent sellers for OFFER generation
		r.addRecentSeller(product, order.TeamName)
	}

	log.Debug().
		Str("product", product).
		Str("side", side).
		Str("clOrdID", order.ClOrdID).
		Str("teamName", order.TeamName).
		Int("buyOrders", len(book.BuyOrders)).
		Int("sellOrders", len(book.SellOrders)).
		Msg("Order added to book")
}

func (r *OrderBookRepository) RemoveOrder(product, side, clOrdID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	book := r.orderBooks[product]
	if book == nil {
		return false
	}

	switch side {
	case "BUY":
		for i, order := range book.BuyOrders {
			if order.ClOrdID == clOrdID {
				book.BuyOrders = append(book.BuyOrders[:i], book.BuyOrders[i+1:]...)
				log.Debug().
					Str("product", product).
					Str("side", side).
					Str("clOrdID", clOrdID).
					Msg("Order removed from book")
				return true
			}
		}
	case "SELL":
		for i, order := range book.SellOrders {
			if order.ClOrdID == clOrdID {
				book.SellOrders = append(book.SellOrders[:i], book.SellOrders[i+1:]...)
				log.Debug().
					Str("product", product).
					Str("side", side).
					Str("clOrdID", clOrdID).
					Msg("Order removed from book")
				return true
			}
		}
	}

	return false
}

func (r *OrderBookRepository) GetBestBid(product string) *domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book := r.orderBooks[product]
	if book == nil || len(book.BuyOrders) == 0 {
		return nil
	}

	// Return highest bid (first in sorted array)
	return book.BuyOrders[0]
}

func (r *OrderBookRepository) GetBestAsk(product string) *domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book := r.orderBooks[product]
	if book == nil || len(book.SellOrders) == 0 {
		return nil
	}

	// Return lowest ask (first in sorted array)
	return book.SellOrders[0]
}

func (r *OrderBookRepository) GetBuyOrders(product string) []*domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book := r.orderBooks[product]
	if book == nil {
		return nil
	}

	// Return copy to avoid race conditions
	orders := make([]*domain.Order, len(book.BuyOrders))
	copy(orders, book.BuyOrders)
	return orders
}

func (r *OrderBookRepository) GetSellOrders(product string) []*domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book := r.orderBooks[product]
	if book == nil {
		return nil
	}

	// Return copy to avoid race conditions
	orders := make([]*domain.Order, len(book.SellOrders))
	copy(orders, book.SellOrders)
	return orders
}

func (r *OrderBookRepository) GetRecentSellers(product string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sellers := r.recentSellers[product]
	if sellers == nil {
		return nil
	}

	// Return copy
	result := make([]string, len(sellers))
	copy(result, sellers)
	return result
}

func (r *OrderBookRepository) addRecentSeller(product, teamName string) {
	sellers := r.recentSellers[product]

	// Remove if already exists
	for i, seller := range sellers {
		if seller == teamName {
			sellers = append(sellers[:i], sellers[i+1:]...)
			break
		}
	}

	// Add to front
	sellers = append([]string{teamName}, sellers...)

	// Keep only last 10 sellers
	if len(sellers) > 10 {
		sellers = sellers[:10]
	}

	r.recentSellers[product] = sellers
}

func (r *OrderBookRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orderBooks = make(map[string]*OrderBook)
	r.recentSellers = make(map[string][]string)

	log.Info().Msg("Order book cleared")
}

func (r *OrderBookRepository) LoadFromDatabase(ctx context.Context, orderRepo domain.OrderRepository) error {
	log.Info().Msg("Loading pending orders from database to order book")

	orders, err := orderRepo.GetPendingOrders(ctx)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear existing order book
	r.orderBooks = make(map[string]*OrderBook)
	r.recentSellers = make(map[string][]string)

	// Add all pending orders
	for _, order := range orders {
		// Get or create order book for product
		if r.orderBooks[order.Product] == nil {
			r.orderBooks[order.Product] = &OrderBook{Product: order.Product}
		}

		book := r.orderBooks[order.Product]

		switch order.Side {
		case "BUY":
			book.BuyOrders = append(book.BuyOrders, order)
		case "SELL":
			book.SellOrders = append(book.SellOrders, order)
			r.addRecentSeller(order.Product, order.TeamName)
		}
	}

	// Sort all order books
	for _, book := range r.orderBooks {
		// Sort buy orders by price DESC
		sort.Slice(book.BuyOrders, func(i, j int) bool {
			if book.BuyOrders[i].Price == nil && book.BuyOrders[j].Price == nil {
				return book.BuyOrders[i].CreatedAt.Before(book.BuyOrders[j].CreatedAt)
			}
			if book.BuyOrders[i].Price == nil {
				return false
			}
			if book.BuyOrders[j].Price == nil {
				return true
			}
			if *book.BuyOrders[i].Price == *book.BuyOrders[j].Price {
				return book.BuyOrders[i].CreatedAt.Before(book.BuyOrders[j].CreatedAt)
			}
			return *book.BuyOrders[i].Price > *book.BuyOrders[j].Price
		})

		// Sort sell orders by price ASC
		sort.Slice(book.SellOrders, func(i, j int) bool {
			if book.SellOrders[i].Price == nil && book.SellOrders[j].Price == nil {
				return book.SellOrders[i].CreatedAt.Before(book.SellOrders[j].CreatedAt)
			}
			if book.SellOrders[i].Price == nil {
				return false
			}
			if book.SellOrders[j].Price == nil {
				return true
			}
			if *book.SellOrders[i].Price == *book.SellOrders[j].Price {
				return book.SellOrders[i].CreatedAt.Before(book.SellOrders[j].CreatedAt)
			}
			return *book.SellOrders[i].Price < *book.SellOrders[j].Price
		})
	}

	totalOrders := 0
	for product, book := range r.orderBooks {
		buyCount := len(book.BuyOrders)
		sellCount := len(book.SellOrders)
		totalOrders += buyCount + sellCount

		log.Info().
			Str("product", product).
			Int("buyOrders", buyCount).
			Int("sellOrders", sellCount).
			Msg("Order book loaded for product")
	}

	log.Info().
		Int("totalOrders", totalOrders).
		Int("products", len(r.orderBooks)).
		Msg("Order book loaded from database")

	return nil
}

// Verify the repository implements the interface
var _ domain.OrderBookRepository = (*OrderBookRepository)(nil)
