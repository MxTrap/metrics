package models

import "errors"

var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrNotFoundMetric    = errors.New("metric not found")
	ErrWrongMetricValue  = errors.New("wrong metric value")
)
