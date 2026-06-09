package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/storage"
	"github.com/spf13/cobra"
)

var (
	saveTTL         string
	saveEncrypted   bool
	saveRemote      bool
	saveDescription string
	saveCategory    string
	saveLang        string
	saveTags        string
)

var saveCmd = &cobra.Command{
	Use:   "save <alias> [content]",
	Short: "Save a snippet locally (persistent, offline-first)",
	Long: `Save a snippet to the local database with an alias.
Use --remote to also sync to the server for cross-device access.
Content can be provided as argument or piped from stdin.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		localDB, err := storage.NewLocalDB(localDBPath())
		if err != nil {
			return fmt.Errorf("open local db: %w", err)
		}
		defer localDB.Close()

		alias := args[0]
		var content string
		if len(args) > 1 {
			content = args[1]
		} else {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			content = string(data)
		}
		if content == "" {
			return fmt.Errorf("no content provided")
		}

		if saveEncrypted {
			pass, err := getMasterPassword()
			if err != nil {
				return fmt.Errorf("get master password: %w", err)
			}
			key := crypto.DeriveKey(pass, cfg.UserID())
			encrypted, err := crypto.Encrypt(content, key)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}
			content = encrypted
		}

		var snippet *models.Snippet
		existing, lookupErr := localDB.GetCachedByAlias(alias)
		if lookupErr == nil && existing != nil {
			snippet, err = localDB.UpdateSnippet(existing.ID, alias, saveDescription, content, saveCategory, saveLang, saveTags)
			if err != nil {
				return fmt.Errorf("update locally: %w", err)
			}
			fmt.Printf("Snippet updated locally (alias: %s, id: %s)\n", snippet.Alias, snippet.ID)
		} else {
			snippet, err = localDB.SaveSnippet(alias, saveDescription, content, saveEncrypted, saveTTL, saveCategory, saveLang, saveTags)
			if err != nil {
				return fmt.Errorf("save locally: %w", err)
			}
			fmt.Printf("Snippet saved locally with id: %s (alias: %s)\n", snippet.ID, snippet.Alias)
		}
		if snippet.ExpiresAt != nil {
			fmt.Printf("  [expires: %s]\n", snippet.ExpiresAt.Format("2006-01-02 15:04:05"))
		}

		// Optionally push to remote
		if saveRemote {
			client := api.NewClient(cfg)
			remoteSnippet, err := client.CreateSnippet(content, alias, saveTTL, saveEncrypted)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to sync to remote: %v\n", err)
			} else {
				fmt.Printf("Also synced to cloud with id: %s\n", remoteSnippet.ID)
			}
		}

		return nil
	},
}

func init() {
	saveCmd.Flags().StringVar(&saveTTL, "ttl", "", "time to live (e.g., 30s, 5m, 1h, 1d)")
	saveCmd.Flags().BoolVar(&saveEncrypted, "encrypt", false, "encrypt content with AES-256-GCM")
	saveCmd.Flags().BoolVar(&saveRemote, "remote", false, "also sync to the remote server")
	saveCmd.Flags().StringVar(&saveDescription, "desc", "", "description for the snippet")
	saveCmd.Flags().StringVar(&saveCategory, "category", "", "category for organizing snippets")
	saveCmd.Flags().StringVar(&saveLang, "lang", "", "language hint for syntax highlighting")
	saveCmd.Flags().StringVar(&saveTags, "tags", "", "comma-separated tags")
	rootCmd.AddCommand(saveCmd)
}
