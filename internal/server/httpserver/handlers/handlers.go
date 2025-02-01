package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricService interface {
	Save(url string) error
	Find(url string) (any, error)
	GetAll() map[string]any
}

type Handler struct {
	service MetricService
}

func NewHandler(service MetricService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h Handler) Save(g *gin.Context) {

	if err := h.service.Save(g.Request.RequestURI); err != nil {
		if errors.Is(err, models.ErrNotFoundMetric) {
			g.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, models.ErrUnknownMetricType) {
			g.Status(http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrWrongMetricValue) {
			g.Status(http.StatusBadRequest)
			return
		}
		g.Status(http.StatusInternalServerError)
		return
	}

	g.Status(http.StatusOK)
}

func (h Handler) Find(g *gin.Context) {
	val, err := h.service.Find(g.Request.RequestURI)
	if err != nil {
		if errors.Is(err, models.ErrNotFoundMetric) {
			g.Status(http.StatusNotFound)
			return
		}
		if errors.Is(err, models.ErrUnknownMetricType) {
			g.Status(http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrWrongMetricValue) {
			g.Status(http.StatusBadRequest)
			return
		}
		g.Status(http.StatusInternalServerError)
		return
	}

	g.String(http.StatusOK, fmt.Sprintf("%v", val))
}

func (h Handler) GetAll(g *gin.Context) {
	g.HTML(http.StatusOK, "index.tmpl", gin.H{
		"metrics": h.service.GetAll(),
	})
}
