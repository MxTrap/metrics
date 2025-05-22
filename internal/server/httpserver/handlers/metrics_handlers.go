// Package handlers предоставляет HTTP-обработчики для управления метриками с использованием фреймворка Gin.
// Определяет MetricsHandler, который интегрируется с MetricService для выполнения операций с метриками.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/gin-gonic/gin"
	"github.com/mailru/easyjson"
	"net/http"
	"strconv"
	"strings"

	"github.com/MxTrap/metrics/internal/server/models"
)

type saver interface {
	Save(ctx context.Context, metrics commonmodels.Metric) error
	SaveAll(ctx context.Context, metrics []commonmodels.Metric) error
}

type getter interface {
	Find(ctx context.Context, metric commonmodels.Metric) (commonmodels.Metric, error)
	GetAll(ctx context.Context) (map[string]any, error)
}

// MetricService определяет интерфейс для операций с метриками, включая сохранение, получение и проверку хранилища.
type MetricService interface {
	saver
	getter
	Ping(ctx context.Context) error
}

// MetricsHandler управляет HTTP-маршрутами и обработчиками для операций с метриками.
// Использует MetricService для взаимодействия с хранилищем и Gin-роутер для обработки запросов.
type MetricsHandler struct {
	router  *gin.Engine
	service MetricService
}

// NewMetricHandler создаёт новый MetricsHandler с указанным MetricService и Gin-роутером.
// Возвращает указатель на инициализированный MetricsHandler.
func NewMetricHandler(service MetricService, router *gin.Engine) *MetricsHandler {
	return &MetricsHandler{
		service: service,
		router:  router,
	}
}

// RegisterRoutes регистрирует HTTP-маршруты для операций с метриками на роутере MetricsHandler.
// Настраивает конечные точки для сохранения, получения и проверки метрик.
func (h MetricsHandler) RegisterRoutes() {
	uri := "/:metricType/:metricName"
	h.router.GET("/value"+uri, h.find)
	h.router.POST("/update/", h.saveJSON)
	h.router.POST(fmt.Sprintf("/update/%s/:metricValue", uri), h.save)
	h.router.POST("/updates/", h.saveAll)
	h.router.POST("/value/", h.findJSON)
	h.router.GET("/", h.getAll)
	h.router.GET("/ping", h.ping)
}

// parseMetric парсит метрику из JSON-данных.
// Возвращает распарсенную метрику или ошибку, если данные некорректны.
func (MetricsHandler) parseMetric(rawData []byte) (commonmodels.Metric, error) {
	m := commonmodels.Metric{}
	err := easyjson.Unmarshal(rawData, &m)
	if err != nil {
		return commonmodels.Metric{}, err
	}
	return m, nil
}

// parseURL извлекает метрику из URL-пути, содержащего указанное ключевое слово.
// Возвращает распарсенную метрику или ошибку, если URL некорректен или значения неверны.
func (MetricsHandler) parseURL(url string, searchWord string) (commonmodels.Metric, error) {
	idx := strings.Index(url, searchWord+"/")
	if idx == -1 {
		return commonmodels.Metric{}, models.ErrNotFoundMetric
	}
	splitedURL := strings.Split(url[idx+len(searchWord)+1:], "/")

	if len(splitedURL) < 2 {
		return commonmodels.Metric{}, models.ErrNotFoundMetric
	}

	metric := commonmodels.Metric{
		ID:    splitedURL[1],
		MType: splitedURL[0],
	}

	if len(splitedURL) == 2 {
		return metric, nil
	}

	val := splitedURL[2]

	if metric.MType == commonmodels.Gauge {
		parseFloat, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return commonmodels.Metric{}, models.ErrWrongMetricValue
		}
		metric.Value = &parseFloat
	}

	if metric.MType == commonmodels.Counter {
		parseInt, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return commonmodels.Metric{}, models.ErrWrongMetricValue
		}
		metric.Delta = &parseInt
	}

	return metric, nil
}

// getMetricValue извлекает значение метрики в зависимости от её типа.
// Возвращает значение метрики как interface{} или nil, если тип не поддерживается.
func (MetricsHandler) getMetricValue(metric commonmodels.Metric) any {
	if metric.MType == commonmodels.Gauge {
		return *metric.Value
	}
	if metric.MType == commonmodels.Counter {
		return *metric.Delta
	}
	return nil
}

// saveAll обрабатывает POST-запросы для сохранения нескольких метрик из JSON-данных.
// Возвращает HTTP 200 при успехе или статус ошибки при неудаче.
func (h MetricsHandler) saveAll(g *gin.Context) {
	rawData, err := g.GetRawData()
	if err != nil {
		g.Status(http.StatusBadRequest)
		return
	}
	m := commonmodels.Metrics{}
	err = json.Unmarshal(rawData, &m)

	if err != nil {
		g.Status(http.StatusBadRequest)
		return
	}

	err = h.service.SaveAll(g, m)
	if err != nil {
		g.Status(http.StatusInternalServerError)
		return
	}
	g.Status(http.StatusOK)
}

// saveJSON обрабатывает POST-запросы для сохранения одной метрики из JSON-данных.
// Возвращает HTTP 200 при успехе или статус ошибки при неудаче.
func (h MetricsHandler) saveJSON(g *gin.Context) {
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

	err = h.service.Save(g, m)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.Status(http.StatusOK)
}

// save обрабатывает POST-запросы для сохранения одной метрики из параметров URL.
// Возвращает HTTP 200 при успехе или статус ошибки при неудаче.
func (h MetricsHandler) save(g *gin.Context) {
	m, err := h.parseURL(g.Request.RequestURI, "update")
	if err == nil {
		err = h.service.Save(g, m)
	}
	if err != nil {
		_ = g.Error(err)
		return
	}

	g.Status(http.StatusOK)
}

// find обрабатывает GET-запросы для получения метрики по типу и имени из параметров URL.
// Возвращает значение метрики в виде строки или статус ошибки при неудаче.
func (h MetricsHandler) find(g *gin.Context) {
	m, err := h.parseURL(g.Request.RequestURI, "value")
	if err == nil {
		m, err = h.service.Find(g, m)
	}

	if err != nil {
		_ = g.Error(err)
		return
	}

	g.String(http.StatusOK, fmt.Sprintf("%v", h.getMetricValue(m)))
}

// findJSON обрабатывает POST-запросы для получения метрики из JSON-данных.
// Возвращает метрику в формате JSON или статус ошибки при неудаче.
func (h MetricsHandler) findJSON(g *gin.Context) {
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
	m, err := h.service.Find(g, metric)
	if err != nil {
		_ = g.Error(err)
		return
	}

	g.JSON(http.StatusOK, m)
}

// getAll обрабатывает GET-запросы для получения всех метрик.
// Возвращает HTML-страницу со всеми метриками или статус ошибки при неудаче.
func (h MetricsHandler) getAll(g *gin.Context) {
	all, err := h.service.GetAll(g)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.HTML(http.StatusOK, "index.tmpl", gin.H{
		"metrics": all,
	})
}

// ping обрабатывает GET-запросы для проверки доступности хранилища метрик.
// Возвращает HTTP 200 при успехе или статус ошибки при неудаче.
func (h MetricsHandler) ping(g *gin.Context) {
	err := h.service.Ping(g)
	if err != nil {
		_ = g.Error(err)
	}
}
