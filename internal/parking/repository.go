package parking

import (
	"context"
	"errors"
	"sort"
	"sync"
)

var (
	ErrSpotNotFound     = errors.New("spot not found")
	ErrSpotNotAvailable = errors.New("spot not available")
)

// Repository describes storage behavior needed by the service.
type Repository interface {
	ListSpots(ctx context.Context) ([]Spot, error)
	ReserveSpot(ctx context.Context, spotID string) (Spot, error)
}

// InMemoryRepository is deterministic storage for local development and tests.
type InMemoryRepository struct {
	mu    sync.RWMutex
	spots map[string]Spot
}

func NewInMemoryRepository(seed []Spot) *InMemoryRepository {
	spots := make(map[string]Spot, len(seed))
	for _, spot := range seed {
		spots[spot.ID] = spot
	}

	return &InMemoryRepository{spots: spots}
}

func (r *InMemoryRepository) ListSpots(ctx context.Context) ([]Spot, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.spots))
	for id := range r.spots {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	spots := make([]Spot, 0, len(ids))
	for _, id := range ids {
		spots = append(spots, r.spots[id])
	}

	return spots, nil
}

func (r *InMemoryRepository) ReserveSpot(ctx context.Context, spotID string) (Spot, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	spot, ok := r.spots[spotID]
	if !ok {
		return Spot{}, ErrSpotNotFound
	}
	if !spot.IsAvailable {
		return Spot{}, ErrSpotNotAvailable
	}

	spot.IsAvailable = false
	r.spots[spotID] = spot

	return spot, nil
}
