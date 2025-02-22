package repository

import (
	"bufio"
	"encoding/json"
	"os"
)

type MetricsFileStorage struct {
	filePath string
}

func NewMetricsFileStorage(filePath string) *MetricsFileStorage {
	return &MetricsFileStorage{
		filePath: filePath,
	}
}

func (s MetricsFileStorage) Save(metrics map[string]any) error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	data, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
func (s MetricsFileStorage) Read() (map[string]any, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			return
		}

	}()

	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	data := scanner.Bytes()

	var res map[string]any
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
