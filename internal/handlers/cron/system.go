package cron

import (
	"ims-pocketbase-baas-starter/pkg/common/jobutils"

	"github.com/pocketbase/pocketbase"
)

// CleanupLogsHandler handles system log cleanup operations
func HandleSystemQueue(app *pocketbase.PocketBase) {
	ctx := jobutils.NewJobExecutionContext(app, "system_queue")
	ctx.LogStart("Starting system queue process operations")

	queues, err := app.FindAllRecords("queues")
	if err != nil {
		ctx.LogError(err, "Error fetching queues data")
	}

	ctx.LogDebug(queues, "fetched queues")

	ctx.LogEnd("System queue process operations completed successfully")
}
