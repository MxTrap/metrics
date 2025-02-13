package handlers

import (
	"errors"
	"fmt"
	common_moodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"net/http"

	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricService interface {
	Save(url string) error
	SaveJSON(metrics common_moodels.Metrics) error
	Find(url string) (any, error)
	FindJSON(metric common_moodels.Metrics) (common_moodels.Metrics, error)
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

func parseMetric(g *gin.Context) (*common_moodels.Metrics, error) {
	rawData, err := g.GetRawData()
	if err != nil {
		return nil, err
	}

	m := common_moodels.Metrics{}
	err = easyjson.Unmarshal(rawData, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (h Handler) SaveJSON(g *gin.Context) {
	m, err := parseMetric(g)
	if err != nil {
		g.Status(http.StatusBadRequest)
		return
	}

	err = h.service.SaveJSON(*m)
	if err == nil {
		g.Status(http.StatusOK)
		return
	}

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
}

func (h Handler) Save(g *gin.Context) {
	err := h.service.Save(g.Request.RequestURI)
	if err == nil {
		g.Status(http.StatusOK)
		return
	}

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
}

func (h Handler) Find(g *gin.Context) {
	val, err := h.service.Find(g.Request.RequestURI)
	if err == nil {
		g.String(http.StatusOK, fmt.Sprintf("%v", val))
		return
	}

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
}

func (h Handler) FindJSON(g *gin.Context) {
	metric, err := parseMetric(g)
	if err != nil {
		g.Status(http.StatusBadRequest)
		return
	}
	val, err := h.service.FindJSON(*metric)
	if err == nil {

		g.JSON(http.StatusOK, val)
		return
	}

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
}

func (h Handler) GetAll(g *gin.Context) {
	g.HTML(http.StatusOK, "index.tmpl", gin.H{
		"metrics": h.service.GetAll(),
	})
}
