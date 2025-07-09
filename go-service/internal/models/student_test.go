package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStudent_FormatName(t *testing.T) {
	tests := []struct {
		name     string
		student  Student
		expected string
	}{
		{
			name:     "Valid name",
			student:  Student{Name: "John Doe"},
			expected: "John Doe",
		},
		{
			name:     "Empty name",
			student:  Student{Name: ""},
			expected: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.student.FormatName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStudent_FormatEmail(t *testing.T) {
	tests := []struct {
		name     string
		student  Student
		expected string
	}{
		{
			name:     "Valid email",
			student:  Student{Email: "john@example.com"},
			expected: "john@example.com",
		},
		{
			name:     "Empty email",
			student:  Student{Email: ""},
			expected: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.student.FormatEmail()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *string
		defaultValue string
		expected     string
	}{
		{
			name:         "Valid string pointer",
			ptr:          stringPtr("test value"),
			defaultValue: "default",
			expected:     "test value",
		},
		{
			name:         "Nil pointer",
			ptr:          nil,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty string pointer",
			ptr:          stringPtr(""),
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeString(tt.ptr, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeInt(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *int
		defaultValue int
		expected     int
	}{
		{
			name:         "Valid int pointer",
			ptr:          intPtr(42),
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "Nil pointer",
			ptr:          nil,
			defaultValue: 10,
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeInt(tt.ptr, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
