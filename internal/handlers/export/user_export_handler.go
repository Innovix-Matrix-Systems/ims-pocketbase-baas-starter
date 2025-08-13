package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ims-pocketbase-baas-starter/pkg/jobutils"
	applogger "ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HandleUserExport processes user export jobs with optimized batch queries
func HandleUserExport(app *pocketbase.PocketBase, jobId string, payload *jobutils.DataProcessingJobPayload) error {
	// Record start time for timeout checking
	startTime := time.Now()

	// Get logger instance
	logger := applogger.GetLogger(app)

	// Fetch all users from the database
	users, err := fetchAllUsers(app)
	if err != nil {
		logger.Error("Failed to fetch users", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	logger.Info("Fetched users for export", "job_id", jobId, "user_count", len(users))

	// Check if we have users to export
	if len(users) == 0 {
		logger.Warn("No users found to export", "job_id", jobId)
		return fmt.Errorf("no users found to export")
	}

	// Check timeout before CSV conversion
	if payload.Options.Timeout > 0 && time.Since(startTime).Seconds() > float64(payload.Options.Timeout) {
		logger.Warn("User export timeout during CSV conversion", "job_id", jobId, "elapsed", time.Since(startTime))
		return fmt.Errorf("export operation timed out")
	}

	// Convert users to CSV
	csvData, err := convertUsersToCSV(app, users)
	if err != nil {
		logger.Error("Failed to convert users to CSV", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to convert users to CSV: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("users_export_%s.csv", time.Now().Format("20060102_150405"))

	logger.Info("Generated CSV data", "job_id", jobId, "filename", filename, "file_size", len(csvData))

	// Save the export file
	if _, err := jobutils.SaveExportedJobFiles(app, jobId, filename, csvData, len(users)); err != nil {
		logger.Error("Failed to save export file", "job_id", jobId, "error", err)
		return fmt.Errorf("failed to save export file: %w", err)
	}

	logger.Info("User export completed successfully", "job_id", jobId, "filename", filename, "user_count", len(users))

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

// convertUsersToCSV converts user records to CSV format using optimized batch queries
func convertUsersToCSV(app *pocketbase.PocketBase, users []*core.Record) ([]byte, error) {
	// Pre-allocate buffer with estimated size to reduce memory allocations
	estimatedSize := len(users) * 200 // Rough estimate of 200 bytes per user row
	var buf bytes.Buffer
	buf.Grow(estimatedSize)

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

	// Build lookup maps for roles and permissions to avoid N+1 queries
	roleNameMap, err := buildRoleNameMap(app, users)
	if err != nil {
		return nil, fmt.Errorf("failed to build role name map: %w", err)
	}

	permissionSlugMap, err := buildPermissionSlugMap(app, users)
	if err != nil {
		return nil, fmt.Errorf("failed to build permission slug map: %w", err)
	}

	// Write user data using pre-built maps
	for _, user := range users {
		// Get role names using the pre-built map
		roleNames := getRoleNames(app, user, roleNameMap)

		// Get permission slugs using the pre-built map
		permissionSlugs := getPermissionSlugs(app, user, permissionSlugMap)

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

// buildRoleNameMap creates a map of role ID to role name for all users (single batch query)
func buildRoleNameMap(app *pocketbase.PocketBase, users []*core.Record) (map[string]string, error) {
	// Collect all unique role IDs across all users
	roleIdSet := make(map[string]struct{})
	for _, user := range users {
		roleIds := user.GetStringSlice("roles")
		for _, roleId := range roleIds {
			if roleId != "" {
				roleIdSet[roleId] = struct{}{}
			}
		}
	}

	// Convert set to slice
	roleIds := make([]string, 0, len(roleIdSet))
	for roleId := range roleIdSet {
		roleIds = append(roleIds, roleId)
	}

	// Single batch query for all roles
	if len(roleIds) == 0 {
		return make(map[string]string), nil
	}

	roles, err := app.FindRecordsByIds("roles", roleIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}

	// Build the map
	roleNameMap := make(map[string]string, len(roles))
	for _, role := range roles {
		roleNameMap[role.Id] = role.GetString("name")
	}

	return roleNameMap, nil
}

// buildPermissionSlugMap creates a map of permission ID to permission slug for all users (single batch query)
func buildPermissionSlugMap(app *pocketbase.PocketBase, users []*core.Record) (map[string]string, error) {
	// Collect all unique permission IDs across all users
	permissionIdSet := make(map[string]struct{})
	for _, user := range users {
		permissionIds := user.GetStringSlice("permissions")
		for _, permissionId := range permissionIds {
			if permissionId != "" {
				permissionIdSet[permissionId] = struct{}{}
			}
		}
	}

	// Convert set to slice
	permissionIds := make([]string, 0, len(permissionIdSet))
	for permissionId := range permissionIdSet {
		permissionIds = append(permissionIds, permissionId)
	}

	// Single batch query for all permissions
	if len(permissionIds) == 0 {
		return make(map[string]string), nil
	}

	permissions, err := app.FindRecordsByIds("permissions", permissionIds)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %w", err)
	}

	// Build the map
	permissionSlugMap := make(map[string]string, len(permissions))
	for _, permission := range permissions {
		permissionSlugMap[permission.Id] = permission.GetString("slug")
	}

	return permissionSlugMap, nil
}

// getRoleNames extracts role names using optimized batch queries to avoid N+1 problem
func getRoleNames(app *pocketbase.PocketBase, user *core.Record, roleNameMap map[string]string) string {
	roleIds := user.GetStringSlice("roles")
	if len(roleIds) == 0 {
		return ""
	}

	roleNames := make([]string, 0, len(roleIds))
	for _, roleId := range roleIds {
		if name, exists := roleNameMap[roleId]; exists {
			roleNames = append(roleNames, name)
		}
	}

	return strings.Join(roleNames, "; ")
}

// getPermissionSlugs extracts permission slugs using optimized batch queries to avoid N+1 problem
func getPermissionSlugs(app *pocketbase.PocketBase, user *core.Record, permissionSlugMap map[string]string) string {
	permissionIds := user.GetStringSlice("permissions")
	if len(permissionIds) == 0 {
		return ""
	}

	permissionSlugs := make([]string, 0, len(permissionIds))
	for _, permissionId := range permissionIds {
		if slug, exists := permissionSlugMap[permissionId]; exists {
			permissionSlugs = append(permissionSlugs, slug)
		}
	}

	return strings.Join(permissionSlugs, "; ")
}
