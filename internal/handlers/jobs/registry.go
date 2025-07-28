package jobs

import (
	"ims-pocketbase-baas-starter/pkg/jobutils"

	"github.com/pocketbase/pocketbase"
)

// InitializeJobHandlers registers all available job handlers with the job processor
func InitializeJobHandlers(app *pocketbase.PocketBase, processor *jobutils.JobProcessor) error {
	app.Logger().Info("Initializing job handlers")

	// Get the registry from the processor
	registry := processor.GetRegistry()

	// Register email job handler
	emailHandler := NewEmailJobHandler(app)
	if err := registry.Register(emailHandler); err != nil {
		app.Logger().Error("Failed to register email job handler", "error", err)
		return err
	}
	app.Logger().Info("Registered job handler", "type", emailHandler.GetJobType())

	// Register data processing job handler
	dataHandler := NewDataProcessingJobHandler(app)
	if err := registry.Register(dataHandler); err != nil {
		app.Logger().Error("Failed to register data processing job handler", "error", err)
		return err
	}
	app.Logger().Info("Registered job handler", "type", dataHandler.GetJobType())

	// Log all registered handlers
	registeredTypes := registry.ListHandlers()
	app.Logger().Info("Job handler initialization completed",
		"total_handlers", len(registeredTypes),
		"registered_types", registeredTypes)

	return nil
}
