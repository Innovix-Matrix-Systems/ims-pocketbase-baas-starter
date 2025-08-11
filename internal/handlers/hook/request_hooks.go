package hook

import (
	"github.com/pocketbase/pocketbase/core"
)

// HandleRecordListRequest handles record list request events
func HandleRecordListRequest(e *core.RecordsListRequestEvent) error {
	// Log the record list request
	e.App.Logger().Debug("Record list requested",
		"collection", e.Collection.Name,
		"user_ip", e.Request.RemoteAddr,
		"user_agent", e.Request.UserAgent(),
	)

	// Add your custom logic here
	// For example: rate limiting, access logging, custom filtering, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordViewRequest handles record view request events
func HandleRecordViewRequest(e *core.RecordRequestEvent) error {
	// Log the record view request
	e.App.Logger().Debug("Record view requested",
		"collection", e.Collection.Name,
		"record_id", e.Record.Id,
		"user_ip", e.Request.RemoteAddr,
	)

	// Add your custom logic here
	// For example: access logging, view tracking, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordCreateRequest handles record create request events
func HandleRecordCreateRequest(e *core.RecordRequestEvent) error {
	// Log the record create request
	e.App.Logger().Debug("Record create requested",
		"collection", e.Collection.Name,
		"user_ip", e.Request.RemoteAddr,
	)

	// Add your custom logic here
	// For example: validation, rate limiting, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordUpdateRequest handles record update request events
func HandleRecordUpdateRequest(e *core.RecordRequestEvent) error {
	// Log the record update request
	e.App.Logger().Debug("Record update requested",
		"collection", e.Collection.Name,
		"record_id", e.Record.Id,
		"user_ip", e.Request.RemoteAddr,
	)

	// Add your custom logic here
	// For example: change tracking, validation, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordDeleteRequest handles record delete request events
func HandleRecordDeleteRequest(e *core.RecordRequestEvent) error {
	// Log the record delete request
	e.App.Logger().Debug("Record delete requested",
		"collection", e.Collection.Name,
		"record_id", e.Record.Id,
		"user_ip", e.Request.RemoteAddr,
	)

	// Add your custom logic here
	// For example: soft delete, backup before delete, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleUserListRequest handles user-specific list requests
func HandleUserListRequest(e *core.RecordsListRequestEvent) error {
	// This is an example of collection-specific request hook
	e.App.Logger().Debug("User list requested",
		"user_ip", e.Request.RemoteAddr,
		"query_params", e.Request.URL.RawQuery,
	)

	// Add user-specific logic here
	// For example: privacy filtering, access control, etc.

	return e.Next()
}
