package jobs

import (
	"sync"

	"ims-pocketbase-baas-starter/internal/handlers/jobs"
	"ims-pocketbase-baas-starter/pkg/jobutils"

	"github.com/pocketbase/pocketbase"
)

// JobManager manages the global job processor instance
type JobManager struct {
	processor   *jobutils.JobProcessor
	mu          sync.RWMutex
	initialized bool
}

var (
	globalJobManager *JobManager
	once             sync.Once
)

// GetJobManager returns the singleton job manager instance
func GetJobManager() *JobManager {
	once.Do(func() {
		globalJobManager = &JobManager{}
	})
	return globalJobManager
}

// Initialize sets up the job processor with all handlers
// This should be called once during application startup
func (jm *JobManager) Initialize(app *pocketbase.PocketBase) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if jm.initialized {
		app.Logger().Debug("Job manager already initialized, skipping")
		return nil
	}

	app.Logger().Info("Initializing job manager and processors")

	// Create the job processor
	jm.processor = jobutils.NewJobProcessor(app)

	// Initialize and register job handlers
	if err := jobs.InitializeJobHandlers(app, jm.processor); err != nil {
		return err
	}

	jm.initialized = true
	app.Logger().Info("Job manager initialization completed - ready for job processing")

	return nil
}

// GetProcessor returns the initialized job processor
// Returns nil if not initialized
func (jm *JobManager) GetProcessor() *jobutils.JobProcessor {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	if !jm.initialized {
		return nil
	}

	return jm.processor
}

// IsInitialized returns whether the job manager has been initialized
func (jm *JobManager) IsInitialized() bool {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.initialized
}
