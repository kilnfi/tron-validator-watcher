package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFormatting(t *testing.T) {
	t.Parallel()

	tf := &CustomTextFormatter{
		DisableColors:  true,
		DisablePadding: true,
	}

	type field struct {
		key   string
		value string
	}
	testCases := []struct {
		name     string
		field    field
		message  string
		level    logrus.Level
		expected string
	}{
		{
			field: field{
				key:   "test",
				value: "test",
			},
			message:  "success message",
			level:    logrus.InfoLevel,
			expected: "0001-01-01T00:00:00Z [INFO] success message  | test=test\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entry := logrus.WithField(tc.field.key, tc.field.value)
			entry.Message = tc.message
			entry.Level = logrus.InfoLevel
			b, _ := tf.Format(entry)

			require.Equal(t, tc.expected, string(b))
		})
	}
}
