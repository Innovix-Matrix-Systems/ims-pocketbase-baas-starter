package middlewares

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func TestNewAuthMiddleware(t *testing.T) {
	middleware := NewAuthMiddleware()
	if middleware == nil {
		t.Fatal("NewAuthMiddleware() returned nil")
	}
}

func TestRequireAuth(t *testing.T) {
	middleware := NewAuthMiddleware()

	// Test without collection names
	handler := middleware.RequireAuth()
	if handler == nil {
		t.Fatal("RequireAuth() returned nil handler")
	}

	// Verify it returns a hook.Handler
	if _, ok := any(handler).(*hook.Handler[*core.RequestEvent]); !ok {
		t.Fatal("RequireAuth() did not return correct handler type")
	}
}

func TestRequireAuthWithCollections(t *testing.T) {
	middleware := NewAuthMiddleware()

	// Test with collection names
	handler := middleware.RequireAuth("users", "_superusers")
	if handler == nil {
		t.Fatal("RequireAuth() with collections returned nil handler")
	}

	// Verify it returns a hook.Handler
	if _, ok := any(handler).(*hook.Handler[*core.RequestEvent]); !ok {
		t.Fatal("RequireAuth() with collections did not return correct handler type")
	}
}

func TestRequireAuthFunc(t *testing.T) {
	middleware := NewAuthMiddleware()

	// Test function wrapper
	authFunc := middleware.RequireAuthFunc()
	if authFunc == nil {
		t.Fatal("RequireAuthFunc() returned nil function")
	}

	// Test with collection names
	authFuncWithCollections := middleware.RequireAuthFunc("users")
	if authFuncWithCollections == nil {
		t.Fatal("RequireAuthFunc() with collections returned nil function")
	}
}
