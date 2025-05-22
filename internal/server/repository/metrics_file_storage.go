// Package repository предоставляет файловое хранилище для метрик.
// Реализует MetricsFileStorage для сохранения, чтения и управления метриками в файле.
package repository

import (
	"bufio"
	"encoding/json"
	commonmodels "github.com/MxTrap/metrics/internal/common/models"
	"os"
)

type MetricsFileStorage struct {
	filePath string
	file     *os.File
}

// NewMetricsFileStorage создаёт новое файловое хранилище метрик по указанному пути.
// Открывает файл с правами чтения и записи, создавая его при необходимости.
// Возвращает указатель на инициализированный MetricsFileStorage или nil при ошибке.
func NewMetricsFileStorage(filePath string) *MetricsFileStorage {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil
	}
	return &MetricsFileStorage{
		filePath: filePath,
		file:     file,
	}
}

// Save сохраняет метрики в файл в формате JSON.
// Очищает файл перед записью и записывает новые данные.
// Возвращает ошибку при неудаче.
func (s *MetricsFileStorage) Save(metrics map[string]commonmodels.Metric) error {
	data, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	err = s.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = s.file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Read считывает метрики из файла в формате JSON.
// Возвращает карту метрик или пустую карту, если файл пуст.
// Возвращает ошибку при неудаче десериализации.
func (s *MetricsFileStorage) Read() (map[string]commonmodels.Metric, error) {
	var data []byte
	scanner := bufio.NewScanner(s.file)

	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	if len(data) == 0 {
		return map[string]commonmodels.Metric{}, nil
	}
	var res map[string]commonmodels.Metric
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Close закрывает файл хранилища.
// Возвращает ошибку при неудаче закрытия.
func (s *MetricsFileStorage) Close() error {
	err := s.file.Close()
	if err != nil {
		return err
	}
	return nil
}
