package servicemanager

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errTestAdditionalContext = errors.New("additional context")
	errTestDifferent         = errors.New("different error")
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrServiceNotFound",
			err:         ErrServiceNotFound,
			expectedMsg: "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestErrorMatching(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "ErrServiceNotFound matches itself",
			err:      ErrServiceNotFound,
			target:   ErrServiceNotFound,
			expected: true,
		},
		{
			name:     "wrapped ErrServiceNotFound matches",
			err:      errors.Join(ErrServiceNotFound, errTestAdditionalContext),
			target:   ErrServiceNotFound,
			expected: true,
		},
		{
			name:     "different error does not match",
			err:      errTestDifferent,
			target:   ErrServiceNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("ErrServiceNotFound is proper error type", func(t *testing.T) {
		err := ErrServiceNotFound
		assert.NotNil(t, err)
		assert.Implements(t, (*error)(nil), ErrServiceNotFound)
	})
}
