package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
	"github.com/spf13/cobra"
)

var (
	stashTTL         string
	stashEncrypted   bool
	stashDescription string
)

var stashCmd = &cobra.Command{
	Use:   "stash <alias> [content]",
	Short: "Save a snippet to the cloud with alias (convenience shortcut)",
	Long: `Save a named snippet to the remote server. This is a convenience shortcut
for 'send --alias'. Content can be provided as argument or piped from stdin.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

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

		if stashEncrypted {
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

		// Upsert: if alias exists, update (creates version history); otherwise create new
		existing, err := client.GetSnippetByAlias(alias)
		var snippet *models.Snippet
		if err == nil && existing != nil {
			snippet, err = client.UpdateSnippet(existing.ID, content)
			if err != nil {
				return fmt.Errorf("update snippet: %w", err)
			}
			fmt.Printf("Snippet updated (alias: %s, id: %s) — version history saved\n", snippet.Alias, snippet.ID)
		} else {
			snippet, err = client.CreateSnippetFull(content, alias, stashDescription, stashTTL, stashEncrypted)
			if err != nil {
				return fmt.Errorf("stash to cloud: %w", err)
			}
			fmt.Printf("Snippet stashed to cloud with id: %s (alias: %s)\n", snippet.ID, snippet.Alias)
		}

		if snippet.ExpiresAt != nil {
			fmt.Printf("  [expires: %s]\n", snippet.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

func init() {
	stashCmd.Flags().StringVar(&stashTTL, "ttl", "", "time to live (e.g., 30s, 5m, 1h, 1d)")
	stashCmd.Flags().BoolVar(&stashEncrypted, "encrypt", false, "encrypt content with AES-256-GCM")
	stashCmd.Flags().StringVar(&stashDescription, "desc", "", "description for the snippet")
	rootCmd.AddCommand(stashCmd)
}
