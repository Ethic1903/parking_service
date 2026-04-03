package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"parking-service/internal/pkg/models"
)

type SQLRepository struct {
	db     *sql.DB
	driver string
	psql   sq.StatementBuilderType
}

func NewSQLRepository(db *sql.DB, driver string) (*SQLRepository, error) {
	normalized := strings.ToLower(strings.TrimSpace(driver))
	if normalized != "postgres" && normalized != "sqlite" {
		return nil, fmt.Errorf("unsupported db driver %q", driver)
	}

	builder := sq.StatementBuilder.PlaceholderFormat(sq.Question)
	if normalized == "postgres" {
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	}

	return &SQLRepository{db: db, driver: normalized, psql: builder}, nil
}

func (r *SQLRepository) SeedSpots(ctx context.Context, spots []models.Spot) error {
	if len(spots) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := r.psql.RunWith(tx)
	for _, spot := range spots {
		query := b.Insert("parking_spots").
			Columns("id", "location", "vehicle_type", "price_per_hour", "is_available").
			Values(spot.ID, spot.Location, spot.VehicleType, spot.PricePerHour, spot.IsAvailable).
			Suffix("ON CONFLICT(id) DO UPDATE SET location = excluded.location, vehicle_type = excluded.vehicle_type, price_per_hour = excluded.price_per_hour")

		sqlQuery, args, err := query.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, sqlQuery, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLRepository) ListSpots(ctx context.Context, filter models.SearchFilter) ([]models.Spot, error) {
	query := r.psql.Select("id", "location", "vehicle_type", "price_per_hour", "is_available").
		From("parking_spots").
		Where(sq.Eq{"is_available": true}).
		OrderBy("id")

	if location := strings.TrimSpace(filter.Location); location != "" {
		query = query.Where("LOWER(location) = LOWER(?)", location)
	}
	if vehicleType := strings.TrimSpace(filter.VehicleType); vehicleType != "" {
		query = query.Where("LOWER(vehicle_type) = LOWER(?)", vehicleType)
	}
	if filter.MaxPricePerHour > 0 {
		query = query.Where(sq.LtOrEq{"price_per_hour": filter.MaxPricePerHour})
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spots := make([]models.Spot, 0)
	for rows.Next() {
		var spot models.Spot
		if err := rows.Scan(&spot.ID, &spot.Location, &spot.VehicleType, &spot.PricePerHour, &spot.IsAvailable); err != nil {
			return nil, err
		}
		spots = append(spots, spot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return spots, nil
}

func (r *SQLRepository) BookSpot(ctx context.Context, req models.BookRequest, createdAtUTC time.Time) (models.Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Booking{}, err
	}
	defer tx.Rollback()

	var (
		pricePerHour int
		isAvailable  bool
	)

	b := r.psql.RunWith(tx)

	selectQuery := b.Select("price_per_hour", "is_available").
		From("parking_spots").
		Where(sq.Eq{"id": req.SpotID})
	selectSQL, selectArgs, err := selectQuery.ToSql()
	if err != nil {
		return models.Booking{}, err
	}

	if err := tx.QueryRowContext(ctx, selectSQL, selectArgs...).Scan(&pricePerHour, &isAvailable); err != nil {
		if err == sql.ErrNoRows {
			return models.Booking{}, ErrSpotNotFound
		}
		return models.Booking{}, err
	}
	if !isAvailable {
		return models.Booking{}, ErrSpotNotAvailable
	}

	updateQuery := b.Update("parking_spots").
		Set("is_available", false).
		Where(sq.Eq{"id": req.SpotID, "is_available": true})
	updateSQL, updateArgs, err := updateQuery.ToSql()
	if err != nil {
		return models.Booking{}, err
	}

	res, err := tx.ExecContext(ctx, updateSQL, updateArgs...)
	if err != nil {
		return models.Booking{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return models.Booking{}, err
	}
	if affected == 0 {
		return models.Booking{}, ErrSpotNotAvailable
	}

	durationHours := int(math.Ceil(req.To.Sub(req.From).Hours()))
	if durationHours < 1 {
		durationHours = 1
	}

	booking := models.Booking{
		ID:         uuid.NewString(),
		SpotID:     req.SpotID,
		UserID:     strings.TrimSpace(req.UserID),
		From:       req.From,
		To:         req.To,
		TotalPrice: durationHours * pricePerHour,
		CreatedAt:  createdAtUTC,
	}

	insertQuery := b.Insert("bookings").
		Columns("id", "spot_id", "user_id", "from_ts", "to_ts", "total_price", "created_at").
		Values(booking.ID, booking.SpotID, booking.UserID, booking.From.UTC(), booking.To.UTC(), booking.TotalPrice, booking.CreatedAt)
	insertSQL, insertArgs, err := insertQuery.ToSql()
	if err != nil {
		return models.Booking{}, err
	}

	if _, err := tx.ExecContext(ctx, insertSQL, insertArgs...); err != nil {
		return models.Booking{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Booking{}, err
	}

	return booking, nil
}
