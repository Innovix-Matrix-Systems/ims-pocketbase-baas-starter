package app

import (
	"ims-pocketbase-baas-starter/pkg/common"
	"testing"
)

// TestAppCreation verifies that the app can be created without errors
func TestAppCreation(t *testing.T) {
	pbApp := NewApp()
	if pbApp == nil {
		t.Fatal("Expected app.NewApp() to return a non-nil app")
	}

	// Verify basic app components are initialized
	if pbApp.Settings() == nil {
		t.Fatal("Expected app to have settings configured")
	}

	if pbApp.OnServe() == nil {
		t.Fatal("Expected OnServe hook to be registered")
	}
}

// TestProtectedCollectionsConfiguration tests the protected collections are correctly defined
func TestProtectedCollectionsConfiguration(t *testing.T) {
	expectedCollections := []string{"users", "roles", "permissions"}

	if len(common.ProtectedCollections) != len(expectedCollections) {
		t.Errorf("Expected %d protected collections, got %d",
			len(expectedCollections), len(common.ProtectedCollections))
	}

	// Verify each expected collection is present
	collectionMap := make(map[string]bool)
	for _, collection := range common.ProtectedCollections {
		collectionMap[collection] = true
	}

	for _, expected := range expectedCollections {
		if !collectionMap[expected] {
			t.Errorf("Expected protected collection '%s' not found in %v",
				expected, common.ProtectedCollections)
		}
	}
}

// TestMiddlewareRegistration tests that the middleware is properly registered
func TestMiddlewareRegistration(t *testing.T) {
	pbApp := NewApp()

	// Get the OnServe hook
	onServeHook := pbApp.OnServe()
	if onServeHook == nil {
		t.Fatal("Expected OnServe hook to be registered")
	}

	// The hook should have at least one handler (our middleware)
	// We can't directly access the handlers, but we can verify the hook exists
	// and that creating the app doesn't panic
}
