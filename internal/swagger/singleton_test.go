package swagger

import (
	"sync"
	"testing"

	"github.com/pocketbase/pocketbase"
)

func TestSingletonPattern(t *testing.T) {
	// Reset singleton before test
	GetInstance().Reset()

	// Create a mock PocketBase app
	app := pocketbase.New()

	// Test that multiple calls return the same instance
	gen1 := InitializeGenerator(app)
	gen2 := InitializeGenerator(app)
	gen3 := GetGlobalGenerator()

	if gen1 != gen2 {
		t.Error("InitializeGenerator should return the same instance on multiple calls")
	}

	if gen1 != gen3 {
		t.Error("GetGlobalGenerator should return the same instance as InitializeGenerator")
	}

	if gen2 != gen3 {
		t.Error("All generator instances should be the same")
	}
}

func TestSingletonThreadSafety(t *testing.T) {
	// Reset singleton before test
	GetInstance().Reset()

	app := pocketbase.New()
	var wg sync.WaitGroup
	generators := make([]*Generator, 10)

	// Launch 10 goroutines to initialize the generator concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			generators[index] = InitializeGenerator(app)
		}(i)
	}

	wg.Wait()

	// All generators should be the same instance
	firstGen := generators[0]
	for i, gen := range generators {
		if gen != firstGen {
			t.Errorf("Generator at index %d is different from the first generator", i)
		}
	}
}

func TestSingletonReset(t *testing.T) {
	// Initialize singleton
	app := pocketbase.New()
	gen1 := InitializeGenerator(app)

	// Reset singleton
	GetInstance().Reset()

	// Initialize again - should be a new instance
	gen2 := InitializeGenerator(app)

	if gen1 == gen2 {
		t.Error("After reset, InitializeGenerator should return a new instance")
	}
}

func TestSingletonIsInitialized(t *testing.T) {
	// Reset singleton before test
	GetInstance().Reset()

	// Should not be initialized initially
	if GetInstance().IsInitialized() {
		t.Error("Singleton should not be initialized initially")
	}

	// Initialize
	app := pocketbase.New()
	InitializeGenerator(app)

	// Should be initialized now
	if !GetInstance().IsInitialized() {
		t.Error("Singleton should be initialized after InitializeGenerator call")
	}
}
