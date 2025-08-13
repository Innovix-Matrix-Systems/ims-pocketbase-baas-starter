package jobutils

import (
	"fmt"
	"time"

	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// SaveExportedJobFiles saves file data to the export_files collection
// This is a utility function that can be used by various export handlers
func SaveExportedJobFiles(app *pocketbase.PocketBase, jobId, filename string, fileData []byte, recordCount int) (*core.Record, error) {
	// Find the export_files collection
	collection, err := app.FindCollectionByNameOrId("export_files")
	if err != nil {
		return nil, fmt.Errorf("export_files collection not found: %w", err)
	}

	// Create a new record
	record := core.NewRecord(collection)

	// Get expiration days from environment variable (default: 30 days)
	expirationDays := common.GetEnvInt("EXPORT_FILE_EXPIRATION_DAYS", 30)
	expirationDate := time.Now().AddDate(0, 0, expirationDays)

	log := logger.GetLogger(app)
	log.Debug("Setting export file expiration",
		"job_id", jobId,
		"expiration_days", expirationDays,
		"expires_at", expirationDate.Format(time.RFC3339))

	// Set the basic fields
	record.Set("job_id", jobId)
	record.Set("user_id", "") // This should be set to the user who requested the export if available
	record.Set("record_count", recordCount)
	record.Set("expires_at", expirationDate)

	// Create a filesystem.File from the file data
	file, err := filesystem.NewFileFromBytes(fileData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file from data: %w", err)
	}

	// Set the file field using PocketBase's file handling
	record.Set("file", file)

	// Save the record (this will automatically handle file upload)
	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to create export_files record: %w", err)
	}

	return record, nil
}

// SaveExportedJobFilesWithUserId saves file data to the export_files collection with a specific user ID
// This variant allows specifying the user who requested the export
func SaveExportedJobFilesWithUserId(app *pocketbase.PocketBase, jobId, userId, filename string, fileData []byte, recordCount int) (*core.Record, error) {
	// Find the export_files collection
	collection, err := app.FindCollectionByNameOrId("export_files")
	if err != nil {
		return nil, fmt.Errorf("export_files collection not found: %w", err)
	}

	// Create a new record
	record := core.NewRecord(collection)

	// Get expiration days from environment variable (default: 30 days)
	expirationDays := common.GetEnvInt("EXPORT_FILE_EXPIRATION_DAYS", 30)
	expirationDate := time.Now().AddDate(0, 0, expirationDays)

	app.Logger().Debug("Setting export file expiration",
		"job_id", jobId,
		"user_id", userId,
		"expiration_days", expirationDays,
		"expires_at", expirationDate.Format(time.RFC3339))

	// Set the basic fields
	record.Set("job_id", jobId)
	record.Set("user_id", userId)
	record.Set("record_count", recordCount)
	record.Set("expires_at", expirationDate)

	// Create a filesystem.File from the file data
	file, err := filesystem.NewFileFromBytes(fileData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file from data: %w", err)
	}

	// Set the file field using PocketBase's file handling
	record.Set("file", file)

	// Save the record (this will automatically handle file upload)
	if err := app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to create export_files record: %w", err)
	}

	return record, nil
}
