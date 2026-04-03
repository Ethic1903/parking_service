package repository

import (
	"context"
	"errors"
	"time"

	"parking-service/internal/pkg/models"
)

var (
	ErrSpotNotFound     = errors.New("spot not found")
	ErrSpotNotAvailable = errors.New("spot not available")
)

// Repository describes storage behavior needed by the service.
type Repository interface {
	ListSpots(ctx context.Context, filter models.SearchFilter) ([]models.Spot, error)
	BookSpot(ctx context.Context, req models.BookRequest, createdAtUTC time.Time) (models.Booking, error)
	SeedSpots(ctx context.Context, spots []models.Spot) error
}
