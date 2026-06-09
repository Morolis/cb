package cmd

import (
	"fmt"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [id|alias]",
	Short: "Get a snippet by ID or alias (checks local first, then remote)",
	Long:  `Retrieve a snippet. If no argument given, returns the most recent snippet.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)
		localDB, err := storage.NewLocalDB(localDBPath())
		if err != nil {
			return fmt.Errorf("open local db: %w", err)
		}
		defer localDB.Close()

		resolver := resolve.NewResolver(localDB, client)

		var snippet *models.Snippet
		var errGet error

		if len(args) == 0 {
			snippet, _, errGet = resolver.GetLatest()
		} else {
			id := args[0]
			if models.IsLocalID(id) {
				snippet, _, errGet = resolver.GetByID(id)
			} else {
				snippet, _, errGet = resolver.GetByAlias(id)
				if errGet != nil {
					snippet, _, errGet = resolver.GetByID(id)
				}
			}
		}

		if errGet != nil {
			return fmt.Errorf("get snippet: %w", errGet)
		}

		if snippet.Encrypted {
			pass, passErr := getMasterPassword()
			if passErr != nil {
				return fmt.Errorf("get master password: %w", passErr)
			}
			key := crypto.DeriveKey(pass, cfg.UserID())
			decrypted, decErr := crypto.Decrypt(snippet.Content, key)
			if decErr != nil {
				return fmt.Errorf("decrypt: %w", decErr)
			}
			snippet.Content = decrypted
		}

		fmt.Print(snippet.Content)
		if len(snippet.Content) > 0 && snippet.Content[len(snippet.Content)-1] != '\n' {
			fmt.Println()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
