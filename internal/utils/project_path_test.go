package utils

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectPath(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	assert.True(t, ok, "runtime.Caller should return valid info")

	expectedPath := filepath.Join(filepath.Dir(testFile), "../..")
	expectedPath, err := filepath.Abs(expectedPath)
	assert.NoError(t, err, "filepath.Abs should not return error")

	result := GetProjectPath()

	resultAbs, err := filepath.Abs(result)
	assert.NoError(t, err, "filepath.Abs should not return error")

	assert.Equal(t, expectedPath, resultAbs, "Project path should match expected")

	assert.False(t, strings.Contains(result, ".."), "Result path should not contain '..'")
	assert.True(t, filepath.IsAbs(resultAbs), "Result should be an absolute path")
}
