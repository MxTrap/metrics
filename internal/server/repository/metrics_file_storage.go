package repository

import (
	"bufio"
	"encoding/json"
	common_models "github.com/MxTrap/metrics/internal/common/models"
	"os"
)

type MetricsFileStorage struct {
	filePath string
	file     *os.File
}

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

func (s MetricsFileStorage) Save(metrics map[string]common_models.Metrics) error {
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
func (s *MetricsFileStorage) Read() (map[string]common_models.Metrics, error) {

	var data []byte
	scanner := bufio.NewScanner(s.file)

	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	var res map[string]common_models.Metrics
	err := json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *MetricsFileStorage) Close() error {
	err := s.file.Close()
	if err != nil {
		return err
	}
	return nil
}
