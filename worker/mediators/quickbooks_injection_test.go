package mediators

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEscapeQuickBooksString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"safe string", "safe string"},
		{"O'Reilly", "O''Reilly"},
		{"' OR '1'='1", "'' OR ''1''=''1"},
		{"Normal Date 2023-01-01", "Normal Date 2023-01-01"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := escapeQuickBooksString(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
