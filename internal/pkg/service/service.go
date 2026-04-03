package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"parking-service/internal/pkg/models"
	"parking-service/internal/pkg/repository"
)

var (
	ErrInvalidTimeRange = errors.New("invalid time range")
	ErrUserIDRequired   = errors.New("user id is required")
)

// Service provides search and booking operations.
type Service struct {
	repo repository.Repository
	now  func() time.Time
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) SearchAvailableSpots(ctx context.Context, filter models.SearchFilter) ([]models.Spot, error) {
	return s.repo.ListSpots(ctx, filter)
}

func (s *Service) BookSpot(ctx context.Context, req models.BookRequest) (models.Booking, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return models.Booking{}, ErrUserIDRequired
	}
	if !req.To.After(req.From) {
		return models.Booking{}, ErrInvalidTimeRange
	}

	req.UserID = strings.TrimSpace(req.UserID)
	return s.repo.BookSpot(ctx, req, s.now().UTC())
}
