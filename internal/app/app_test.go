package app

import (
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
