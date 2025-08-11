package hook

import (
	"github.com/pocketbase/pocketbase/core"
)

// HandleRealtimeConnect handles realtime connection events
func HandleRealtimeConnect(e *core.RealtimeConnectRequestEvent) error {
	// Log the realtime connection
	e.App.Logger().Debug("Realtime client connected",
		"client_id", e.Client.Id(),
	)

	// Add your custom logic here
	// For example: connection tracking, authentication, rate limiting, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRealtimeSubscribe handles realtime subscription events
func HandleRealtimeSubscribe(e *core.RealtimeSubscribeRequestEvent) error {
	// Log the realtime subscription
	e.App.Logger().Debug("Realtime subscription created",
		"client_id", e.Client.Id(),
		"subscriptions", len(e.Subscriptions),
	)

	// Add your custom logic here
	// For example: subscription validation, access control, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRealtimeMessage handles realtime message events
func HandleRealtimeMessage(e *core.RealtimeMessageEvent) error {
	// Log the realtime message
	e.App.Logger().Debug("Realtime message sent",
		"type", e.Message.Name,
		"data_size", len(e.Message.Data),
	)

	// Add your custom logic here
	// For example: message filtering, transformation, logging, etc.

	// Continue with the execution chain
	return e.Next()
}
