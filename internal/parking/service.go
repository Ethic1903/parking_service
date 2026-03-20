package parking

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrUserIDRequired   = errors.New("user id is required")
)

// Service provides search and booking operations.
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) SearchAvailableSpots(ctx context.Context, filter SearchFilter) ([]Spot, error) {
	spots, err := s.repo.ListSpots(ctx)
	if err != nil {
		return nil, err
	}

	location := strings.ToLower(strings.TrimSpace(filter.Location))
	vehicleType := strings.ToLower(strings.TrimSpace(filter.VehicleType))

	result := make([]Spot, 0, len(spots))
	for _, spot := range spots {
		if !spot.IsAvailable {
			continue
		}
		if location != "" && strings.ToLower(spot.Location) != location {
			continue
		}
		if vehicleType != "" && strings.ToLower(spot.VehicleType) != vehicleType {
			continue
		}
		if filter.MaxPricePerHour > 0 && spot.PricePerHour > filter.MaxPricePerHour {
			continue
		}

		result = append(result, spot)
	}

	return result, nil
}

func (s *Service) BookSpot(ctx context.Context, req BookRequest) (Booking, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return Booking{}, ErrUserIDRequired
	}
	if !req.To.After(req.From) {
		return Booking{}, ErrInvalidTimeRange
	}

	spot, err := s.repo.ReserveSpot(ctx, req.SpotID)
	if err != nil {
		return Booking{}, err
	}

	durationHours := int(math.Ceil(req.To.Sub(req.From).Hours()))
	if durationHours < 1 {
		durationHours = 1
	}

	booking := Booking{
		ID:         uuid.NewString(),
		SpotID:     spot.ID,
		UserID:     strings.TrimSpace(req.UserID),
		From:       req.From,
		To:         req.To,
		TotalPrice: durationHours * spot.PricePerHour,
		CreatedAt:  s.now().UTC(),
	}

	return booking, nil
}
