package cmd

import (
	"fmt"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [id|alias] [version-id]",
	Short: "Rollback a snippet to a previous version",
	Long:  `Restore a snippet to a specific version. The current content is saved as a new version before rollback.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)
		localDB, err := storage.NewLocalDB(localDBPath())
		if err != nil {
			return fmt.Errorf("open local db: %w", err)
		}
		defer localDB.Close()

		resolver := resolve.NewResolver(localDB, client)

		id := args[0]
		snippet, source, err := resolver.GetByAlias(id)
		if err != nil {
			snippet, source, err = resolver.GetByID(id)
		}
		if err != nil {
			return fmt.Errorf("resolve snippet: %w", err)
		}

		if source == "local" {
			return fmt.Errorf("rollback is only available for remote snippets")
		}

		var versionID uint
		if _, err := fmt.Sscanf(args[1], "%d", &versionID); err != nil {
			return fmt.Errorf("invalid version ID: %s", args[1])
		}

		updated, err := client.Rollback(snippet.ID, versionID)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Println("Rolled back successfully.")
		fmt.Printf("Content:\n%s\n", updated.Content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
