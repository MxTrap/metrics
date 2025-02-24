package service

import (
	"fmt"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"time"
)

type storageGetter interface {
	GetAll() map[string]common_models.Metrics
	Find(metric string) (common_models.Metrics, bool)
}

type storageSaver interface {
	Save(metrics common_models.Metrics)
	SaveAll(metrics map[string]common_models.Metrics)
}

type Storage interface {
	storageGetter
	storageSaver
}

type FileStorage interface {
	Save(metrics map[string]common_models.Metrics) error
	Read() (map[string]common_models.Metrics, error)
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

func (s *StorageService) Save(metrics common_models.Metrics) {
	s.storage.Save(metrics)
	if s.saveInterval == 0 {
		s.saveToFile()
	}
}
func (s *StorageService) Find(metric string) (common_models.Metrics, bool) {
	return s.storage.Find(metric)
}
func (s *StorageService) GetAll() map[string]any {
	dst := map[string]any{}
	metrics := s.storage.GetAll()
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
	return dst
}

func (s *StorageService) saveToFile() {
	err := s.fileStorage.Save(s.storage.GetAll())
	if err != nil {
		return
	}
}

func (s *StorageService) Start() error {
	if s.restore {
		read, err := s.fileStorage.Read()
		fmt.Println(read, err)
		if err != nil {
			return err
		}
		s.storage.SaveAll(read)
	}

	if s.saveInterval > 0 {
		s.ticker = time.NewTicker(time.Duration(s.saveInterval) * time.Second)
		go func() {
			for range s.ticker.C {
				s.saveToFile()
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
