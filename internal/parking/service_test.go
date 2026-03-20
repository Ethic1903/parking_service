package parking

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSearchAvailableSpotsFilters(t *testing.T) {
	repo := NewInMemoryRepository([]Spot{
		{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true},
		{ID: "A-102", Location: "center", VehicleType: "car", PricePerHour: 220, IsAvailable: true},
		{ID: "B-201", Location: "center", VehicleType: "bike", PricePerHour: 90, IsAvailable: true},
		{ID: "C-301", Location: "center", VehicleType: "car", PricePerHour: 140, IsAvailable: false},
	})
	service := NewService(repo)

	spots, err := service.SearchAvailableSpots(context.Background(), SearchFilter{
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
	repo := NewInMemoryRepository([]Spot{
		{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true},
	})
	service := NewService(repo)
	fixedNow := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	from := time.Date(2026, time.March, 20, 13, 0, 0, 0, time.UTC)
	to := from.Add(90 * time.Minute)

	booking, err := service.BookSpot(context.Background(), BookRequest{
		SpotID: "A-101",
		UserID: "student-1",
		From:   from,
		To:     to,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if booking.TotalPrice != 300 {
		t.Fatalf("expected total price 300, got %d", booking.TotalPrice)
	}
	if !booking.CreatedAt.Equal(fixedNow) {
		t.Fatalf("expected createdAt %v, got %v", fixedNow, booking.CreatedAt)
	}

	_, err = service.BookSpot(context.Background(), BookRequest{
		SpotID: "A-101",
		UserID: "student-2",
		From:   from,
		To:     to,
	})
	if !errors.Is(err, ErrSpotNotAvailable) {
		t.Fatalf("expected ErrSpotNotAvailable, got %v", err)
	}
}

func TestBookSpotValidatesInput(t *testing.T) {
	repo := NewInMemoryRepository([]Spot{{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true}})
	service := NewService(repo)
	base := time.Date(2026, time.March, 20, 13, 0, 0, 0, time.UTC)

	_, err := service.BookSpot(context.Background(), BookRequest{
		SpotID: "A-101",
		UserID: "",
		From:   base,
		To:     base.Add(time.Hour),
	})
	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}

	_, err = service.BookSpot(context.Background(), BookRequest{
		SpotID: "A-101",
		UserID: "student-1",
		From:   base,
		To:     base,
	})
	if !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}
