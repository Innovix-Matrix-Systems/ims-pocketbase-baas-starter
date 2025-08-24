package hook

import (
	"ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase/core"
)

// HandleRealtimeConnect handles realtime connection events
func HandleRealtimeConnect(e *core.RealtimeConnectRequestEvent) error {
	// Log the realtime connection
	if log := logger.FromApp(e.App); log != nil {
		log.Debug("Realtime client connected",
			"client_id", e.Client.Id(),
		)
	}

	// Add your custom logic here
	// For example: connection tracking, authentication, rate limiting, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRealtimeSubscribe handles realtime subscription events
func HandleRealtimeSubscribe(e *core.RealtimeSubscribeRequestEvent) error {
	// Log the realtime subscription
	if log := logger.FromApp(e.App); log != nil {
		log.Debug("Realtime subscription created",
			"client_id", e.Client.Id(),
			"subscriptions", len(e.Subscriptions),
		)
	}

	// Add your custom logic here
	// For example: subscription validation, access control, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRealtimeMessage handles realtime message events
func HandleRealtimeMessage(e *core.RealtimeMessageEvent) error {
	// Log the realtime message
	if log := logger.FromApp(e.App); log != nil {
		log.Debug("Realtime message sent",
			"type", e.Message.Name,
			"data_size", len(e.Message.Data),
		)
	}

	// Add your custom logic here
	// For example: message filtering, transformation, logging, etc.

	// Continue with the execution chain
	return e.Next()
}
