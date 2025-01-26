package handlers

import (
	"errors"
	"net/http"

	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricsSaver interface {
	Save(url string) error
}

type Handler struct {
	service MetricsSaver
}

func NewHandler(service MetricsSaver) *Handler {
	return &Handler{
		service: service,
	}
}

func (h Handler) Save(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if err := h.service.Save(r.RequestURI); err != nil {
		if errors.Is(err, models.ErrNotFoundMetric) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, models.ErrUnknownMetricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrWrongMetricValue) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
