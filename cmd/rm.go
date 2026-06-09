package cmd

import (
	"fmt"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/spf13/cobra"
)

var rmSource string

var rmCmd = &cobra.Command{
	Use:   "rm <id|alias>",
	Short: "Delete a snippet",
	Long:  `Delete a snippet by its ID or alias. Routes to local or remote based on ID prefix.`,
	Args:  cobra.ExactArgs(1),
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

		// If source is explicitly specified, use that
		if rmSource == "local" {
			if models.IsLocalID(id) {
				if err := localDB.DeleteCached(id); err != nil {
					return fmt.Errorf("delete local snippet: %w", err)
				}
			} else {
				if err := localDB.DeleteByAlias(id); err != nil {
					return fmt.Errorf("delete local snippet: %w", err)
				}
			}
			fmt.Printf("Local snippet %s deleted.\n", id)
			return nil
		}

		if rmSource == "remote" {
			snippet, err := client.GetSnippetByAlias(id)
			if err == nil {
				id = snippet.ID
			}
			if err := client.DeleteSnippet(id); err != nil {
				return fmt.Errorf("delete remote snippet: %w", err)
			}
			fmt.Printf("Remote snippet %s deleted.\n", args[0])
			return nil
		}

		// Auto-route based on ID prefix
		if models.IsLocalID(id) {
			if err := resolver.Delete(id); err != nil {
				return fmt.Errorf("delete: %w", err)
			}
		} else {
			if err := resolver.DeleteByAlias(id); err != nil {
				// Try by ID
				if err := resolver.Delete(id); err != nil {
					return fmt.Errorf("delete: %w", err)
				}
			}
		}

		fmt.Printf("Snippet %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	rmCmd.Flags().StringVar(&rmSource, "source", "", "force delete from: local, remote (default: auto-detect)")
	rootCmd.AddCommand(rmCmd)
}
