package models

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
