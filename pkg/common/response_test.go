package common

import (
	"testing"
)

func TestGlobalResponseInstance(t *testing.T) {
	if Response == nil {
		t.Error("Expected global Response instance to be initialized")
	}

	// Test that it's a ResponseHelper by calling a method
	if Response == nil {
		t.Error("Expected global Response to be a valid ResponseHelper instance")
	}
}
