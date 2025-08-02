package jobutils

import (
	"testing"
)

func TestValidateJobPayload(t *testing.T) {
	tests := []struct {
		name        string
		payload     map[string]any
		expectError bool
	}{
		{
			name:        "nil payload",
			payload:     nil,
			expectError: true,
		},
		{
			name:        "missing type field",
			payload:     map[string]any{},
			expectError: true,
		},
		{
			name: "valid payload with type",
			payload: map[string]any{
				"type": "test_job",
			},
			expectError: false,
		},
		{
			name: "valid payload with type and data",
			payload: map[string]any{
				"type": "test_job",
				"data": map[string]any{"key": "value"},
			},
			expectError: false,
		},
		{
			name: "invalid type field - not string",
			payload: map[string]any{
				"type": 123,
			},
			expectError: true,
		},
		{
			name: "empty type field",
			payload: map[string]any{
				"type": "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJobPayload(tt.payload)
			if tt.expectError && err == nil {
				t.Errorf("expected error for %q, but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.name, err)
			}
		})
	}
}

func TestNewJobRegistry(t *testing.T) {
	registry := NewJobRegistry()
	if registry == nil {
		t.Error("NewJobRegistry should not return nil")
	}

	handlers := registry.ListHandlers()
	if len(handlers) != 0 {
		t.Errorf("new registry should have 0 handlers, got %d", len(handlers))
	}
}
