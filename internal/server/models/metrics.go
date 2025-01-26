package models

import (
	"errors"
	"strconv"
)

type Action struct {
	ParseFunc func(str string) (any, error)
	SaveFunc  func(storage map[string]any, key string, val any) error
}

var AcceptableMetricTypes = map[string]Action{
	"gauge": {
		ParseFunc: func(str string) (any, error) {
			return strconv.ParseFloat(str, 64)
		},
		SaveFunc: func(storage map[string]any, key string, val any) error {
			storage[key] = val
			return nil
		},
	},
	"counter": {
		ParseFunc: func(str string) (any, error) {
			return strconv.ParseInt(str, 10, 64)
		},
		SaveFunc: func(storage map[string]any, key string, val any) error {
			storedVal, ok := storage[key].(int64)
			if !ok {
				storedVal = 0
			}
			convertedVal, ok := val.(int64)
			if !ok {
				return errors.New("can't convert stored value")
			}
			storage[key] = storedVal + convertedVal

			return nil
		}},
}
