package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"parking-service/internal/parking"
)

type Handler struct {
	service *parking.Service
}

func NewHandler(service *parking.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/api/v1/spots", h.handleGetSpots)
	mux.HandleFunc("/api/v1/bookings", h.handleCreateBooking)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleGetSpots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	maxPrice := 0
	if rawMaxPrice := r.URL.Query().Get("maxPrice"); rawMaxPrice != "" {
		parsed, err := strconv.Atoi(rawMaxPrice)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "maxPrice must be an integer")
			return
		}
		maxPrice = parsed
	}

	spots, err := h.service.SearchAvailableSpots(r.Context(), parking.SearchFilter{
		Location:        r.URL.Query().Get("location"),
		VehicleType:     r.URL.Query().Get("vehicleType"),
		MaxPricePerHour: maxPrice,
	})
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to search spots")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]any{"items": spots})
}

func (h *Handler) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SpotID string `json:"spotId"`
		UserID string `json:"userId"`
		From   string `json:"from"`
		To     string `json:"to"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	from, err := time.Parse(time.RFC3339, payload.From)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "from must be RFC3339")
		return
	}
	to, err := time.Parse(time.RFC3339, payload.To)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "to must be RFC3339")
		return
	}

	booking, err := h.service.BookSpot(r.Context(), parking.BookRequest{
		SpotID: payload.SpotID,
		UserID: payload.UserID,
		From:   from,
		To:     to,
	})
	if err != nil {
		switch {
		case errors.Is(err, parking.ErrSpotNotFound):
			h.writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, parking.ErrSpotNotAvailable):
			h.writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, parking.ErrInvalidTimeRange), errors.Is(err, parking.ErrUserIDRequired):
			h.writeError(w, http.StatusBadRequest, err.Error())
		default:
			h.writeError(w, http.StatusInternalServerError, "failed to book spot")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, booking)
}

func (h *Handler) writeError(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSON(w, statusCode, map[string]string{"error": message})
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
