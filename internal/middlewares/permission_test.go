package middlewares

import (
	"testing"
)

// TestHasPermission tests the HasPermission function with various scenarios
func TestHasPermission(t *testing.T) {
	pm := NewPermissionMiddleware()

	tests := []struct {
		name             string
		userPermissions  []string
		checkPermissions []string
		expected         bool
	}{
		{
			name:             "user has required permission",
			userPermissions:  []string{"user.create", "user.view"},
			checkPermissions: []string{"user.create"},
			expected:         true,
		},
		{
			name:             "user has one of multiple required permissions",
			userPermissions:  []string{"user.create", "user.view"},
			checkPermissions: []string{"user.delete", "user.view"},
			expected:         true,
		},
		{
			name:             "user doesn't have required permission",
			userPermissions:  []string{"user.create", "user.view"},
			checkPermissions: []string{"user.delete"},
			expected:         false,
		},
		{
			name:             "empty user permissions",
			userPermissions:  []string{},
			checkPermissions: []string{"user.create"},
			expected:         false,
		},
		{
			name:             "empty check permissions",
			userPermissions:  []string{"user.create"},
			checkPermissions: []string{},
			expected:         false,
		},
		{
			name:             "both empty",
			userPermissions:  []string{},
			checkPermissions: []string{},
			expected:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.HasPermission(tt.userPermissions, tt.checkPermissions)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
