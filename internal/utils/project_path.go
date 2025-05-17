package utils

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
)

func GetProjectPath() string {
	return filepath.Join(filepath.Dir(b), "../..")
}
