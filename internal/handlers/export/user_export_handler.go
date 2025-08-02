package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ims-pocketbase-baas-starter/pkg/jobutils"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HandleUserExport processes user export jobs by:
func HandleUserExport(app *pocketbase.PocketBase, jobId string, payload *jobutils.DataProcessingJobPayload) error {
	// Record start time for timeout checking
	startTime := time.Now()

	// Fetch all users from the database
	users, err := fetchAllUsers(app)
	if err != nil {
		app.Logger().Error("Failed to fetch users", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	app.Logger().Info("Fetched users for export", "job_id", jobId, "user_count", len(users))

	// Check if we have users to export
	if len(users) == 0 {
		app.Logger().Warn("No users found to export", "job_id", jobId)
		return fmt.Errorf("no users found to export")
	}

	// Check timeout before CSV conversion
	if payload.Options.Timeout > 0 && time.Since(startTime).Seconds() > float64(payload.Options.Timeout) {
		app.Logger().Warn("User export timeout during CSV conversion", "job_id", jobId, "elapsed", time.Since(startTime))
		return fmt.Errorf("export operation timed out")
	}

	// Convert users to CSV
	csvData, err := convertUsersToCSV(app, users)
	if err != nil {
		app.Logger().Error("Failed to convert users to CSV", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to convert users to CSV: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("users_export_%s.csv", time.Now().Format("20060102_150405"))

	app.Logger().Info("Generated CSV data", "job_id", jobId, "filename", filename, "file_size", len(csvData))

	// Save to export_files collection
	exportRecord, err := jobutils.SaveExportedJobFiles(app, jobId, filename, csvData, len(users))
	if err != nil {
		app.Logger().Error("Failed to save export file", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to save export file: %w", err)
	}

	app.Logger().Info("User export completed successfully",
		"job_id", jobId,
		"export_record_id", exportRecord.Id,
		"filename", filename,
		"user_count", len(users),
		"file_size", len(csvData))

	return nil
}

// fetchAllUsers retrieves all users from the users collection
func fetchAllUsers(app *pocketbase.PocketBase) ([]*core.Record, error) {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return nil, fmt.Errorf("users collection not found: %w", err)
	}

	// Fetch all users
	records, err := app.FindRecordsByFilter(
		collection,
		"",         // no filter - get all users
		"-created", // sort by created date descending
		0,          // no limit
		0,          // no offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	return records, nil
}

// convertUsersToCSV converts user records to CSV format
func convertUsersToCSV(app *pocketbase.PocketBase, users []*core.Record) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"ID",
		"Email",
		"Name",
		"Email Visibility",
		"Verified",
		"Is Active",
		"Roles",
		"Permissions",
		"Created",
		"Updated",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write user data
	for _, user := range users {
		// Get role names
		roleNames := getRoleNames(app, user)

		// Get permission slugs
		permissionSlugs := getPermissionSlugs(app, user)

		row := []string{
			user.Id,
			user.GetString("email"),
			user.GetString("name"),
			strconv.FormatBool(user.GetBool("emailVisibility")),
			strconv.FormatBool(user.GetBool("verified")),
			strconv.FormatBool(user.GetBool("is_active")),
			roleNames,
			permissionSlugs,
			user.GetDateTime("created").Time().Format(time.RFC3339),
			user.GetDateTime("updated").Time().Format(time.RFC3339),
		}

		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write user row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// getRoleNames extracts role names from roles relation by fetching the actual role records
func getRoleNames(app *pocketbase.PocketBase, user *core.Record) string {
	roleIds := user.GetStringSlice("roles")
	if len(roleIds) == 0 {
		return ""
	}

	var roleNames []string
	for _, roleId := range roleIds {
		if role, err := app.FindRecordById("roles", roleId); err == nil {
			roleNames = append(roleNames, role.GetString("name"))
		}
	}

	return strings.Join(roleNames, "; ")
}

// getPermissionSlugs extracts permission slugs from permissions relation by fetching the actual permission records
func getPermissionSlugs(app *pocketbase.PocketBase, user *core.Record) string {
	permissionIds := user.GetStringSlice("permissions")
	if len(permissionIds) == 0 {
		return ""
	}

	var permissionSlugs []string
	for _, permissionId := range permissionIds {
		if permission, err := app.FindRecordById("permissions", permissionId); err == nil {
			permissionSlugs = append(permissionSlugs, permission.GetString("slug"))
		}
	}

	return strings.Join(permissionSlugs, "; ")
}
