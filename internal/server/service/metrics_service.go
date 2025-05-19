// Package service предоставляет сервис для управления метриками, включая их сохранение, получение и синхронизацию с хранилищем.
// Реализует MetricsService, который взаимодействует с хранилищем метрик и файловой системой.
package service

import (
	"context"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"github.com/MxTrap/metrics/internal/server/models"
	"time"
)

type storageGetter interface {
	GetAll(ctx context.Context) (map[string]commonmodels.Metric, error)
	Find(ctx context.Context, metric string) (commonmodels.Metric, error)
}

type storageSaver interface {
	Save(ctx context.Context, metrics commonmodels.Metric) error
	SaveAll(ctx context.Context, metrics map[string]commonmodels.Metric) error
}

type Storage interface {
	storageGetter
	storageSaver
	Ping(ctx context.Context) error
}

type fileStorage interface {
	Save(metrics map[string]commonmodels.Metric) error
	Read() (map[string]commonmodels.Metric, error)
	Close() error
}

type MetricsService struct {
	fileStorage  fileStorage
	storage      Storage
	saveInterval int
	restore      bool
	ticker       *time.Ticker
}

// NewMetricsService создаёт новый MetricsService с указанным файловым хранилищем, хранилищем, интервалом сохранения и флагом восстановления.
// Возвращает указатель на инициализированный MetricsService.
func NewMetricsService(fileStorage fileStorage, storage Storage, saveInterval int, restore bool) *MetricsService {
	return &MetricsService{
		fileStorage:  fileStorage,
		storage:      storage,
		saveInterval: saveInterval,
		restore:      restore,
	}
}

func (*MetricsService) validateMetric(metricType string) bool {
	_, ok := models.MetricTypes[metricType]
	return ok
}

// SaveAll сохраняет массив метрик в хранилище.
// Агрегирует метрики типа Counter и выполняет синхронное сохранение в файл, если saveInterval равен 0.
// Возвращает ошибку при неудаче.
func (s *MetricsService) SaveAll(ctx context.Context, metrics []commonmodels.Metric) error {
	m := make(map[string]commonmodels.Metric, len(metrics))
	for _, metric := range metrics {
		if !s.validateMetric(metric.MType) {
			continue
		}
		if val, ok := m[metric.ID]; metric.MType == commonmodels.Counter && ok {
			*metric.Delta = *metric.Delta + *val.Delta
		}

		m[metric.ID] = metric
	}

	err := s.storage.SaveAll(ctx, m)
	if err != nil {
		return err
	}
	if s.saveInterval == 0 {
		err := s.saveToFile(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save сохраняет одну метрику в хранилище.
// Выполняет синхронное сохранение в файл, если saveInterval равен 0.
// Возвращает ошибку при неверном типе метрики, отсутствии значения или неудаче сохранения.
func (s *MetricsService) Save(ctx context.Context, metric commonmodels.Metric) error {
	if !s.validateMetric(metric.MType) {
		return models.ErrUnknownMetricType
	}

	if metric.Delta == nil && metric.Value == nil {
		return models.ErrWrongMetricValue
	}

	err := s.storage.Save(ctx, metric)
	if err != nil {
		return err
	}
	if s.saveInterval == 0 {
		err := s.saveToFile(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Find получает метрику по её идентификатору и типу.
// Возвращает метрику или ошибку при неверном типе или отсутствии метрики.
func (s *MetricsService) Find(ctx context.Context, metric commonmodels.Metric) (commonmodels.Metric, error) {
	if !s.validateMetric(metric.MType) {
		return commonmodels.Metric{}, models.ErrUnknownMetricType
	}

	val, err := s.storage.Find(ctx, metric.ID)
	if err != nil {
		return commonmodels.Metric{}, models.ErrNotFoundMetric
	}

	return val, nil
}

// GetAll возвращает все метрики из хранилища в виде карты с их значениями.
// Возвращает карту или ошибку при неудаче.
func (s *MetricsService) GetAll(ctx context.Context) (map[string]any, error) {
	dst := map[string]any{}
	metrics, err := s.storage.GetAll(ctx)
	if err != nil {
		return dst, err
	}
	for k, v := range metrics {
		var val any
		if v.Delta != nil {
			val = *v.Delta
		}
		if v.Value != nil {
			val = *v.Value
		}
		dst[k] = val
	}
	return dst, nil
}

func (s *MetricsService) saveToFile(ctx context.Context) error {
	all, err := s.storage.GetAll(ctx)
	if err != nil {
		return err
	}
	err = s.fileStorage.Save(all)
	if err != nil {
		return err
	}
	return nil
}

// Ping проверяет доступность хранилища.
// Возвращает ошибку, если хранилище недоступно.
func (s *MetricsService) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

// Start запускает сервис, восстанавливая метрики из файла, если restore=true, и начиная периодическое сохранение, если saveInterval>0.
// Возвращает ошибку при неудаче восстановления или сохранения.
func (s *MetricsService) Start(ctx context.Context) error {
	if s.restore {
		read, err := s.fileStorage.Read()
		if err != nil {
			return err
		}
		err = s.storage.SaveAll(ctx, read)
		if err != nil {
			return err
		}
	}

	if s.saveInterval > 0 {
		s.ticker = time.NewTicker(time.Duration(s.saveInterval) * time.Second)
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-s.ticker.C:
					err := s.saveToFile(ctx)
					if err != nil {
						return
					}
				}
			}
		}(ctx)
	}
	return nil
}

// Stop останавливает периодическое сохранение, сохраняет метрики в файл и закрывает файловое хранилище.
func (s *MetricsService) Stop() {
	s.ticker.Stop()
	err := s.saveToFile(context.Background())
	if err != nil {
		return
	}
	err = s.fileStorage.Close()
	if err != nil {
		return
	}
}
