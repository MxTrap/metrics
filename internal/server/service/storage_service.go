package service

import (
	"context"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"time"
)

type storageGetter interface {
	GetAll(ctx context.Context) (map[string]commonmodels.Metrics, error)
	Find(ctx context.Context, metric string) (commonmodels.Metrics, error)
}

type storageSaver interface {
	Save(ctx context.Context, metrics commonmodels.Metrics) error
	SaveAll(ctx context.Context, metrics map[string]commonmodels.Metrics) error
}

type Storage interface {
	storageGetter
	storageSaver
	Ping(ctx context.Context) error
}

type FileStorage interface {
	Save(metrics map[string]commonmodels.Metrics) error
	Read() (map[string]commonmodels.Metrics, error)
	Close() error
}

type StorageService struct {
	fileStorage  FileStorage
	storage      Storage
	saveInterval int
	restore      bool
	ticker       *time.Ticker
}

func NewStorageService(fileStorage FileStorage, storage Storage, saveInterval int, restore bool) *StorageService {
	return &StorageService{
		fileStorage:  fileStorage,
		storage:      storage,
		saveInterval: saveInterval,
		restore:      restore,
	}
}

func (s *StorageService) Save(ctx context.Context, metrics commonmodels.Metrics) error {
	err := s.storage.Save(ctx, metrics)
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
func (s *StorageService) Find(ctx context.Context, metric string) (commonmodels.Metrics, error) {
	return s.storage.Find(ctx, metric)
}
func (s *StorageService) GetAll(ctx context.Context) (map[string]any, error) {
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

func (s *StorageService) saveToFile(ctx context.Context) error {
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
func (s *StorageService) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *StorageService) Start(ctx context.Context) error {
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
		go func() {
			for range s.ticker.C {
				err := s.saveToFile(ctx)
				if err != nil {
					return
				}
			}
		}()
	}
	return nil
}

func (s *StorageService) Stop() {
	s.ticker.Stop()
	err := s.fileStorage.Close()
	if err != nil {
		return
	}
}
