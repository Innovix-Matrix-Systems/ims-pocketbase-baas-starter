package hooks

import (
	"ims-pocketbase-baas-starter/pkg/metrics"
	"testing"

	"github.com/pocketbase/pocketbase"
)

func TestRegisterHooks(t *testing.T) {
	// Create a test PocketBase app
	app := pocketbase.New()

	// Test that RegisterHooks doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterHooks panicked: %v", r)
		}
	}()

	// Register hooks
	RegisterHooks(app)

	// Verify that hooks are registered by checking if the hook exists
	// Note: PocketBase doesn't provide direct access to registered hooks,
	// so we mainly test that registration doesn't fail

	// Test individual registration functions
	registerRecordHooks(app)
	registerCollectionHooks(app)
	registerRequestHooks(app)
	registerMailerHooks(app)
	registerRealtimeHooks(app)
}

func TestHookRegistrationFunctions(t *testing.T) {
	app := pocketbase.New()

	// Test each registration function individually
	tests := []struct {
		name string
		fn   func(*pocketbase.PocketBase)
	}{
		{"registerRecordHooks", registerRecordHooks},
		{"registerCollectionHooks", registerCollectionHooks},
		{"registerRequestHooks", registerRequestHooks},
		{"registerMailerHooks", registerMailerHooks},
		{"registerRealtimeHooks", registerRealtimeHooks},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panicked: %v", tt.name, r)
				}
			}()

			tt.fn(app)
		})
	}
}

func TestHooksWithMetricsInstrumentation(t *testing.T) {
	app := pocketbase.New()

	// Initialize metrics with no-op provider for testing
	metrics.InitializeProvider(metrics.Config{
		Provider: metrics.ProviderDisabled,
		Enabled:  false,
	})

	// Test that hooks with metrics instrumentation don't panic during registration
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Hooks with metrics instrumentation panicked: %v", r)
		}
	}()

	// Test the instrumented hooks specifically
	registerRecordHooks(app) // Contains user_create_settings instrumentation
	registerMailerHooks(app) // Contains email operation instrumentation

	// Reset metrics for cleanup
	metrics.Reset()
}
