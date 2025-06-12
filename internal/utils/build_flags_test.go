package utils

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatFlagValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "N/A",
		},
		{
			name:     "Non-empty string",
			input:    "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "String with spaces",
			input:    "  test  ",
			expected: "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFlagValue(tt.input)
			assert.Equal(t, tt.expected, result, "formatFlagValue should return expected value")
		})
	}
}

func TestPrintBuildFlags(t *testing.T) {
	tests := []struct {
		name           string
		buildVersion   string
		buildDate      string
		buildCommit    string
		expectedOutput string
	}{
		{
			name:           "All values provided",
			buildVersion:   "v1.0.0",
			buildDate:      "2023-10-01",
			buildCommit:    "abc123",
			expectedOutput: "Build version: v1.0.0\nBuild date: 2023-10-01\nBuild commit: abc123\n",
		},
		{
			name:           "All values empty",
			buildVersion:   "",
			buildDate:      "",
			buildCommit:    "",
			expectedOutput: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		},
		{
			name:           "Mixed values",
			buildVersion:   "v2.0.0",
			buildDate:      "",
			buildCommit:    "def456",
			expectedOutput: "Build version: v2.0.0\nBuild date: N/A\nBuild commit: def456\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Перехватываем вывод
			originalStdout := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err, "failed to create pipe")
			os.Stdout = w

			// Вызываем функцию
			PrintBuildFlags(tt.buildDate, tt.buildCommit, tt.buildVersion)

			// Закрываем запись и восстанавливаем stdout
			w.Close()
			os.Stdout = originalStdout

			// Читаем вывод
			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			require.NoError(t, err, "failed to read from pipe")

			// Проверяем результат
			assert.Equal(t, tt.expectedOutput, buf.String(), "PrintBuildFlags output should match expected")
		})
	}
}
