package utils

import "time"

func Retry(fn func() error, retryCount int) error {
	var err error
	for i := 0; i < retryCount; i++ {
		err = fn()
		if err == nil {
			break
		}

		if i < retryCount {
			time.Sleep(time.Duration(1+2*i) * time.Second)
		}
	}
	return err
}
