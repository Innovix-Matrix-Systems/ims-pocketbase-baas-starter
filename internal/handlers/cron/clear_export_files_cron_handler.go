package cron

import (
	"fmt"
	"time"

	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/cronutils"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HandleClearExportFiles processes cleanup of expired export files
func HandleClearExportFiles(app *pocketbase.PocketBase) {
	ctx := cronutils.NewCronExecutionContext(app, "clear_export_files")
	ctx.LogStart("Starting export files cleanup operations")

	// Get configuration for cleanup batch size
	batchSize := common.GetEnvInt("EXPORT_CLEANUP_BATCH_SIZE", 100) // Process up to 100 expired files per run

	// Find expired export files
	expiredRecords, err := findExpiredExportFiles(app, batchSize)
	if err != nil {
		ctx.LogError(err, "Failed to find expired export files")
		return
	}

	if len(expiredRecords) == 0 {
		ctx.LogDebug("No expired export files found", "No cleanup needed")
		ctx.LogEnd("Export files cleanup completed - no files to clean")
		return
	}

	ctx.LogDebug(fmt.Sprintf("Found %d expired export files to clean up", len(expiredRecords)), "Starting cleanup process")

	// Clean up each expired record
	deletedCount := 0
	errorCount := 0

	for _, record := range expiredRecords {
		if err := deleteExportFileRecord(ctx, app, record); err != nil {
			ctx.LogError(err, fmt.Sprintf("Failed to delete export file record: %s", record.Id))
			errorCount++
			continue
		}
		deletedCount++
	}

	// Log final results
	app.Logger().Info("Export files cleanup batch completed",
		"total_expired", len(expiredRecords),
		"deleted", deletedCount,
		"errors", errorCount,
		"batch_size", batchSize)

	if errorCount > 0 {
		ctx.LogError(fmt.Errorf("cleanup completed with %d errors out of %d records", errorCount, len(expiredRecords)), "Cleanup had errors")
	}

	ctx.LogEnd("Export files cleanup operations completed successfully")
}

// findExpiredExportFiles finds all export file records that have expired
func findExpiredExportFiles(app *pocketbase.PocketBase, batchSize int) ([]*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("export_files")
	if err != nil {
		return nil, fmt.Errorf("export_files collection not found: %w", err)
	}

	// Query for records where expires_at is less than current time
	now := time.Now()
	filter := fmt.Sprintf("expires_at <= '%s'", now.Format(time.RFC3339))

	records, err := app.FindRecordsByFilter(
		collection,
		filter,
		"created", // sort by creation date (oldest first)
		batchSize, // limit to batch size
		0,         // no offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired export files: %w", err)
	}

	return records, nil
}

// deleteExportFileRecord deletes an export file record and its associated file
func deleteExportFileRecord(ctx *cronutils.CronExecutionContext, app *pocketbase.PocketBase, record *core.Record) error {
	recordId := record.Id
	jobId := record.GetString("job_id")
	filename := record.GetString("file")
	expiresAt := record.GetDateTime("expires_at").Time()

	ctx.LogDebug(fmt.Sprintf("Deleting expired export file: record_id=%s, job_id=%s, filename=%s, expired_at=%s",
		recordId, jobId, filename, expiresAt.Format(time.RFC3339)), "Processing expired file")

	// Delete the record (this will automatically delete the associated file)
	if err := app.Delete(record); err != nil {
		return fmt.Errorf("failed to delete export file record %s: %w", recordId, err)
	}

	app.Logger().Info("Deleted expired export file",
		"record_id", recordId,
		"job_id", jobId,
		"filename", filename,
		"expired_at", expiresAt.Format(time.RFC3339))

	return nil
}
