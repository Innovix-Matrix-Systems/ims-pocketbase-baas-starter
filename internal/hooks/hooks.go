package hooks

import (
	"ims-pocketbase-baas-starter/internal/handlers/hook"
	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterHooks registers all custom event hooks
func RegisterHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)
	log.Info("Registering custom event hooks")

	// Register Record hooks
	registerRecordHooks(app)

	// Register Collection hooks
	registerCollectionHooks(app)

	// Register Request hooks
	registerRequestHooks(app)

	// Register Mailer hooks
	registerMailerHooks(app)

	// Register Realtime hooks
	registerRealtimeHooks(app)

	log.Info("Custom event hooks registration completed")
}

// registerRecordHooks registers all record-related event hooks
func registerRecordHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)

	// Example: Log all record creations
	app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
		return hook.HandleRecordCreate(e)
	})

	// Example: Log all record updates
	app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
		return hook.HandleRecordUpdate(e)
	})

	// Example: Handle record deletions
	app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
		return hook.HandleRecordDelete(e)
	})

	// Example: Collection-specific hooks
	// app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
	//     return hook.HandleUserCreate(e)
	// })

	// Example: Additional hook registrations (uncomment to enable)
	// app.OnRecordAfterCreateSuccess().BindFunc(func(e *core.RecordEvent) error {
	//     return hook.HandleAuditLog(e)
	// })

	// app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
	//     return hook.HandleUserWelcomeEmail(e)
	// })

	// app.OnRecordCreateRequest().BindFunc(func(e *core.RecordCreateRequestEvent) error {
	//     return hook.HandleDataValidation(&core.RecordEvent{
	//         App:    e.App,
	//         Record: e.Record,
	//     })
	// })

	// app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
	//     return hook.HandleCacheInvalidation(e)
	// })

	//create user default settings (with metrics instrumentation)
	app.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		// Get the metrics provider instance
		metricsProvider := metrics.GetInstance()

		// Instrument the hook execution with metrics collection
		return metrics.InstrumentHook(metricsProvider, "user_create_settings", func() error {
			return hook.HandleUserCreateSettings(e)
		})
	})

	log.Debug("Record hooks registered")
}

// registerCollectionHooks registers all collection-related event hooks
func registerCollectionHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)

	// Example: Log collection creations
	app.OnCollectionCreate().BindFunc(func(e *core.CollectionEvent) error {
		return hook.HandleCollectionCreate(e)
	})

	// Example: Log collection updates
	app.OnCollectionUpdate().BindFunc(func(e *core.CollectionEvent) error {
		return hook.HandleCollectionUpdate(e)
	})

	log.Debug("Collection hooks registered")
}

// registerRequestHooks registers all request-related event hooks
func registerRequestHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)

	// Example: Log all record list requests
	app.OnRecordsListRequest().BindFunc(func(e *core.RecordsListRequestEvent) error {
		return hook.HandleRecordListRequest(e)
	})

	// Example: Log all record view requests
	app.OnRecordViewRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		return hook.HandleRecordViewRequest(e)
	})

	// Example: Collection-specific request hooks
	// app.OnRecordListRequest("users").BindFunc(func(e *core.RecordListRequestEvent) error {
	//     return hook.HandleUserListRequest(e)
	// })

	log.Debug("Request hooks registered")
}

// registerMailerHooks registers all mailer-related event hooks
func registerMailerHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)

	// Example: Log all email sends (with metrics instrumentation)
	app.OnMailerSend().BindFunc(func(e *core.MailerEvent) error {
		// Get the metrics provider instance
		metricsProvider := metrics.GetInstance()

		// Instrument the email operation with metrics collection
		return metrics.InstrumentEmailOperation(metricsProvider, func() error {
			return hook.HandleMailerSend(e)
		})
	})

	log.Debug("Mailer hooks registered")
}

// registerRealtimeHooks registers all realtime-related event hooks
func registerRealtimeHooks(app *pocketbase.PocketBase) {
	log := logger.GetLogger(app)

	// Example: Log realtime connections
	app.OnRealtimeConnectRequest().BindFunc(func(e *core.RealtimeConnectRequestEvent) error {
		return hook.HandleRealtimeConnect(e)
	})

	// Example: Log realtime disconnections
	app.OnRealtimeSubscribeRequest().BindFunc(func(e *core.RealtimeSubscribeRequestEvent) error {
		return hook.HandleRealtimeSubscribe(e)
	})

	log.Debug("Realtime hooks registered")
}
