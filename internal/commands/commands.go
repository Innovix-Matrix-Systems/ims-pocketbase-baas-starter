package commands

import (
	"ims-pocketbase-baas-starter/internal/handlers/command"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

// Command represents a console command with its configuration
type Command struct {
	ID      string                                                 // Unique identifier for the command
	Use     string                                                 // Command usage string (e.g., "hello", "migrate [type]")
	Short   string                                                 // Short description of the command
	Long    string                                                 // Long description of the command
	Handler func(*pocketbase.PocketBase, *cobra.Command, []string) // Handler function to execute
	Enabled bool                                                   // Whether the command should be registered
}

// RegisterCommands registers all custom console commands with the PocketBase application
// This function follows the same pattern as RegisterCrons, RegisterJobs, and RegisterRoutes
func RegisterCommands(app *pocketbase.PocketBase) {
	if app == nil {
		panic("RegisterCommands: app cannot be nil")
	}

	// Define all custom commands
	commands := []Command{
		{
			ID:      "health",
			Use:     "health",
			Short:   "Perform application health check",
			Long:    "Run a basic health check to verify database connectivity and other core services",
			Handler: command.HandleHealthCheckCommand,
			Enabled: true,
		},
		{
			ID:      "sync-permissions",
			Use:     "sync-permissions",
			Short:   "Sync hardcoded permissions to database",
			Long:    "Syncs all hardcoded permissions defined in the codebase to the database, creating new ones and skipping existing ones",
			Handler: command.HandleSyncPermissionsCommand,
			Enabled: true,
		},
		// Add more commands here as needed:
		// {
		//     ID:      "example",
		//     Use:     "example [arg]",
		//     Short:   "Example command",
		//     Long:    "Detailed description of what this example command does",
		//     Handler: command.HandleExampleCommand,
		//     Enabled: true,
		// },
	}

	// Register enabled commands
	for _, cmd := range commands {
		if !cmd.Enabled {
			continue
		}

		// Create the cobra command
		cobraCmd := &cobra.Command{
			Use:   cmd.Use,
			Short: cmd.Short,
			Long:  cmd.Long,
			Run: func(innerCmd *cobra.Command, args []string) {
				cmd.Handler(app, innerCmd, args)
			},
		}

		// Register the command with PocketBase
		app.RootCmd.AddCommand(cobraCmd)
	}
}
