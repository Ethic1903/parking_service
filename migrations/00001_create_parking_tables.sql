-- +goose Up
CREATE TABLE IF NOT EXISTS parking_spots (
    id TEXT PRIMARY KEY,
    location TEXT NOT NULL,
    vehicle_type TEXT NOT NULL,
    price_per_hour INTEGER NOT NULL,
    is_available BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS bookings (
    id TEXT PRIMARY KEY,
    spot_id TEXT NOT NULL REFERENCES parking_spots(id),
    user_id TEXT NOT NULL,
    from_ts TIMESTAMPTZ NOT NULL,
    to_ts TIMESTAMPTZ NOT NULL,
    total_price INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS parking_spots;
