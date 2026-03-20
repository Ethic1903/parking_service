package parking

import "time"

// Spot describes a parking spot visible to clients.
type Spot struct {
	ID           string `json:"id"`
	Location     string `json:"location"`
	VehicleType  string `json:"vehicleType"`
	PricePerHour int    `json:"pricePerHour"`
	IsAvailable  bool   `json:"isAvailable"`
}

// SearchFilter defines search criteria for available spots.
type SearchFilter struct {
	Location        string
	VehicleType     string
	MaxPricePerHour int
}

// BookRequest contains booking parameters from API handlers.
type BookRequest struct {
	SpotID string
	UserID string
	From   time.Time
	To     time.Time
}

// Booking stores booking result details.
type Booking struct {
	ID         string    `json:"id"`
	SpotID     string    `json:"spotId"`
	UserID     string    `json:"userId"`
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`
	TotalPrice int       `json:"totalPrice"`
	CreatedAt  time.Time `json:"createdAt"`
}
