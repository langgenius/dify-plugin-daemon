package log

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expected    slog.Level
		expectedErr string
	}{
		{
			name:     "empty defaults to info",
			value:    "",
			expected: slog.LevelInfo,
		},
		{
			name:     "uppercase info",
			value:    "INFO",
			expected: slog.LevelInfo,
		},
		{
			name:     "uppercase debug",
			value:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			name:        "invalid level",
			value:       "verbose",
			expectedErr: `invalid LOG_LEVEL "verbose". Valid values are: DEBUG, INFO, WARN, ERROR`,
		},
		{
			name:        "lowercase rejected",
			value:       "info",
			expectedErr: `invalid LOG_LEVEL "info". Valid values are: DEBUG, INFO, WARN, ERROR`,
		},
		{
			name:        "warning alias rejected",
			value:       "WARNING",
			expectedErr: `invalid LOG_LEVEL "WARNING". Valid values are: DEBUG, INFO, WARN, ERROR`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := ParseLevel(tt.value)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, level)
		})
	}
}
