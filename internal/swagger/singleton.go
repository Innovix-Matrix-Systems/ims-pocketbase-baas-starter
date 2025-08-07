package swagger

import (
	"sync"

	"github.com/pocketbase/pocketbase"
)

// SwaggerSingleton manages the global swagger generator instance
type SwaggerSingleton struct {
	generator *Generator
	mu        sync.RWMutex
	once      sync.Once
}

var instance *SwaggerSingleton

// GetInstance returns the singleton instance of SwaggerSingleton
func GetInstance() *SwaggerSingleton {
	if instance == nil {
		instance = &SwaggerSingleton{}
	}
	return instance
}

// Initialize initializes the swagger generator (thread-safe, called only once)
func (s *SwaggerSingleton) Initialize(app *pocketbase.PocketBase) *Generator {
	s.once.Do(func() {
		config := DefaultConfig()
		s.generator = NewGenerator(app, config)
	})
	return s.generator
}

// GetGenerator returns the initialized generator instance
func (s *SwaggerSingleton) GetGenerator() *Generator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generator
}

// IsInitialized checks if the generator has been initialized
func (s *SwaggerSingleton) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generator != nil
}

// Reset resets the singleton (primarily for testing)
func (s *SwaggerSingleton) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.generator = nil
	s.once = sync.Once{}
}
