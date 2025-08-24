package command

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"

	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/permission"
)

// HandleSyncPermissionsCommand syncs hardcoded permissions into the database
func HandleSyncPermissionsCommand(app *pocketbase.PocketBase, cmd *cobra.Command, args []string) {
	log := logger.GetLogger(app)
	log.Info("Starting permission sync process")

	// Get all hardcoded permissions
	hardcodedPermissions := permission.GetAllPermissions()
	log.Info("Found hardcoded permissions", "count", len(hardcodedPermissions))

	// Get the permissions collection
	permissionsCollection, err := app.FindCollectionByNameOrId("permissions")
	if err != nil {
		log.Error("Failed to find permissions collection", "error", err)
		return
	}

	// Track statistics
	var createdCount, skippedCount int

	// Process permissions in batches for better performance
	batchSize := 50
	recordsToCreate := make([]*core.Record, 0, batchSize)

	for _, perm := range hardcodedPermissions {
		// Check if permission already exists
		existingRecord, err := findPermissionBySlug(app, perm.Slug)
		if err != nil {
			log.Error("Error checking permission existence", "slug", perm.Slug, "error", err)
			continue
		}

		if existingRecord != nil {
			log.Debug("Permission already exists, skipping", "slug", perm.Slug)
			skippedCount++
			continue
		}

		// Create new permission record
		record := core.NewRecord(permissionsCollection)
		record.Set("slug", perm.Slug)
		record.Set("name", perm.Name)
		record.Set("description", perm.Description)

		recordsToCreate = append(recordsToCreate, record)

		// Process batch when it reaches batchSize
		if len(recordsToCreate) >= batchSize {
			if err := savePermissionBatch(app, recordsToCreate); err != nil {
				log.Error("Failed to save permission batch", "error", err)
			} else {
				createdCount += len(recordsToCreate)
				log.Info("Saved permission batch", "count", len(recordsToCreate))
			}
			recordsToCreate = make([]*core.Record, 0, batchSize)
		}
	}

	// Process remaining records
	if len(recordsToCreate) > 0 {
		if err := savePermissionBatch(app, recordsToCreate); err != nil {
			log.Error("Failed to save final permission batch", "error", err)
		} else {
			createdCount += len(recordsToCreate)
			log.Info("Saved final permission batch", "count", len(recordsToCreate))
		}
	}

	log.Info("Permission sync completed",
		"created", createdCount,
		"skipped", skippedCount,
		"total_processed", createdCount+skippedCount)
}

// findPermissionBySlug checks if a permission with the given slug already exists
func findPermissionBySlug(app *pocketbase.PocketBase, slug string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"permissions",
		"slug = {:slug}",
		"",
		1,
		0,
		dbx.Params{"slug": slug},
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

// savePermissionBatch saves a batch of permission records
func savePermissionBatch(app *pocketbase.PocketBase, records []*core.Record) error {
	for _, record := range records {
		if err := app.Save(record); err != nil {
			return fmt.Errorf("failed to save permission %s: %w", record.GetString("slug"), err)
		}
	}
	return nil
}
