package handlers

import (
	"fmt"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"net/http"
	"strconv"
	"strings"

	"github.com/MxTrap/metrics/internal/server/models"
)

type MetricService interface {
	Save(metrics common_models.Metrics) error
	Find(metric common_models.Metrics) (common_models.Metrics, error)
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

func (Handler) parseMetric(rawData []byte) (common_models.Metrics, error) {
	m := common_models.Metrics{}
	err := easyjson.Unmarshal(rawData, &m)
	if err != nil {
		return common_models.Metrics{}, err
	}
	return m, nil
}

func (Handler) parseURL(url string, searchWord string) (common_models.Metrics, error) {
	idx := strings.Index(url, searchWord+"/")
	if idx == -1 {
		return common_models.Metrics{}, models.ErrNotFoundMetric
	}
	splitedURL := strings.Split(url[idx+len(searchWord)+1:], "/")

	if len(splitedURL) < 2 {
		return common_models.Metrics{}, models.ErrNotFoundMetric
	}

	metric := common_models.Metrics{
		ID:    splitedURL[1],
		MType: splitedURL[0],
	}

	if len(splitedURL) == 2 {
		return metric, nil
	}

	val := splitedURL[2]

	if metric.MType == common_models.Gauge {
		parseFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return common_models.Metrics{}, models.ErrWrongMetricValue
		}
		metric.Value = &parseFloat
	}

	if metric.MType == common_models.Counter {
		parseInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return common_models.Metrics{}, models.ErrWrongMetricValue
		}
		metric.Delta = &parseInt
	}

	return metric, nil
}

func (Handler) getMetricValue(metric common_models.Metrics) any {
	if metric.MType == common_models.Gauge {
		return *metric.Value
	}
	if metric.MType == common_models.Counter {
		return *metric.Delta
	}
	return nil
}

func (h Handler) SaveJSON(g *gin.Context) {
	rawData, err := g.GetRawData()
	if err != nil {
		g.Status(http.StatusBadRequest)
		return
	}
	m, err := h.parseMetric(rawData)

	if err != nil {
		_ = g.Error(err)
		return
	}

	err = h.service.Save(m)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.Status(http.StatusOK)
}

func (h Handler) Save(g *gin.Context) {
	m, err := h.parseURL(g.Request.RequestURI, "update")
	if err == nil {
		err = h.service.Save(m)
	}
	if err != nil {
		_ = g.Error(err)
		return
	}

	g.Status(http.StatusOK)
}

func (h Handler) Find(g *gin.Context) {
	m, err := h.parseURL(g.Request.RequestURI, "value")
	if err == nil {
		m, err = h.service.Find(m)
	}

	if err != nil {
		_ = g.Error(err)
		return
	}

	g.String(http.StatusOK, fmt.Sprintf("%v", h.getMetricValue(m)))
}

func (h Handler) FindJSON(g *gin.Context) {
	rawData, err := g.GetRawData()
	if err != nil {
		_ = g.Error(err)
		return
	}
	metric, err := h.parseMetric(rawData)
	if err != nil {
		_ = g.Error(err)
		return
	}
	m, err := h.service.Find(metric)
	if err != nil {
		_ = g.Error(err)
		return
	}

	g.JSON(http.StatusOK, h.getMetricValue(m))

}

func (h Handler) GetAll(g *gin.Context) {
	g.HTML(http.StatusOK, "index.tmpl", gin.H{
		"metrics": h.service.GetAll(),
	})
}
