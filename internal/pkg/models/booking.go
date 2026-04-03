package models

import "time"

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
