package repository

import "parking-service/internal/pkg/models"

func DefaultSeedSpots() []models.Spot {
	return []models.Spot{
		{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true},
		{ID: "A-102", Location: "center", VehicleType: "car", PricePerHour: 170, IsAvailable: true},
		{ID: "B-201", Location: "airport", VehicleType: "car", PricePerHour: 220, IsAvailable: true},
		{ID: "C-301", Location: "station", VehicleType: "bike", PricePerHour: 80, IsAvailable: true},
	}
}
