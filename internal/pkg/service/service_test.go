package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"parking-service/internal/pkg/models"
	"parking-service/internal/pkg/repository"
)

type testRepository struct {
	spots map[string]models.Spot
}

func newTestRepository(seed []models.Spot) *testRepository {
	spots := make(map[string]models.Spot, len(seed))
	for _, spot := range seed {
		spots[spot.ID] = spot
	}
	return &testRepository{spots: spots}
}

func (r *testRepository) ListSpots(_ context.Context, filter models.SearchFilter) ([]models.Spot, error) {
	result := make([]models.Spot, 0, len(r.spots))
	location := strings.ToLower(strings.TrimSpace(filter.Location))
	vehicleType := strings.ToLower(strings.TrimSpace(filter.VehicleType))

	for _, spot := range r.spots {
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

func (r *testRepository) BookSpot(_ context.Context, req models.BookRequest, createdAtUTC time.Time) (models.Booking, error) {
	spot, ok := r.spots[req.SpotID]
	if !ok {
		return models.Booking{}, repository.ErrSpotNotFound
	}
	if !spot.IsAvailable {
		return models.Booking{}, repository.ErrSpotNotAvailable
	}

	spot.IsAvailable = false
	r.spots[spot.ID] = spot

	durationHours := int(math.Ceil(req.To.Sub(req.From).Hours()))
	if durationHours < 1 {
		durationHours = 1
	}

	return models.Booking{
		ID:         uuid.NewString(),
		SpotID:     req.SpotID,
		UserID:     req.UserID,
		From:       req.From,
		To:         req.To,
		TotalPrice: durationHours * spot.PricePerHour,
		CreatedAt:  createdAtUTC,
	}, nil
}

func (r *testRepository) SeedSpots(_ context.Context, _ []models.Spot) error {
	return nil
}

func TestSearchAvailableSpotsFilters(t *testing.T) {
	repo := newTestRepository([]models.Spot{
		{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true},
		{ID: "A-102", Location: "center", VehicleType: "car", PricePerHour: 220, IsAvailable: true},
		{ID: "B-201", Location: "center", VehicleType: "bike", PricePerHour: 90, IsAvailable: true},
		{ID: "C-301", Location: "center", VehicleType: "car", PricePerHour: 140, IsAvailable: false},
	})
	svc := NewService(repo)

	spots, err := svc.SearchAvailableSpots(context.Background(), models.SearchFilter{
		Location:        "center",
		VehicleType:     "car",
		MaxPricePerHour: 180,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spots) != 1 {
		t.Fatalf("expected 1 spot, got %d", len(spots))
	}
	if spots[0].ID != "A-101" {
		t.Fatalf("expected A-101, got %s", spots[0].ID)
	}
}

func TestBookSpotReservesAvailability(t *testing.T) {
	repo := newTestRepository([]models.Spot{{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true}})
	svc := NewService(repo)
	fixedNow := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixedNow }

	from := time.Date(2026, time.March, 20, 13, 0, 0, 0, time.UTC)
	to := from.Add(90 * time.Minute)

	booking, err := svc.BookSpot(context.Background(), models.BookRequest{SpotID: "A-101", UserID: "student-1", From: from, To: to})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if booking.TotalPrice != 300 {
		t.Fatalf("expected total price 300, got %d", booking.TotalPrice)
	}
	if !booking.CreatedAt.Equal(fixedNow) {
		t.Fatalf("expected createdAt %v, got %v", fixedNow, booking.CreatedAt)
	}

	_, err = svc.BookSpot(context.Background(), models.BookRequest{SpotID: "A-101", UserID: "student-2", From: from, To: to})
	if !errors.Is(err, repository.ErrSpotNotAvailable) {
		t.Fatalf("expected ErrSpotNotAvailable, got %v", err)
	}
}

func TestBookSpotValidatesInput(t *testing.T) {
	repo := newTestRepository([]models.Spot{{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true}})
	svc := NewService(repo)
	base := time.Date(2026, time.March, 20, 13, 0, 0, 0, time.UTC)

	_, err := svc.BookSpot(context.Background(), models.BookRequest{SpotID: "A-101", UserID: "", From: base, To: base.Add(time.Hour)})
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}

	_, err = svc.BookSpot(context.Background(), models.BookRequest{SpotID: "A-101", UserID: "student-1", From: base, To: base})
	if !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}
