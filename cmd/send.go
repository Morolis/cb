package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/Morolis/cb/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	sendAlias       string
	sendID          string
	sendTTL         string
	sendEncrypted   bool
	sendDescription string
	sendVars        []string
)

var sendCmd = &cobra.Command{
	Use:   "send [content]",
	Short: "Send text to the cloud clipboard (cross-device sync)",
	Long: `Send text content to the remote server for cross-device synchronization.
Content can be provided as argument or piped from stdin.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		var content string
		if len(args) > 0 {
			content = args[0]
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

		// Variable substitution: replace {{.KEY}} with values from --var or environment
		if len(sendVars) > 0 {
			for _, v := range sendVars {
				parts := strings.SplitN(v, "=", 2)
				if len(parts) == 2 {
					placeholder := "{{." + parts[0] + "}}"
					content = strings.ReplaceAll(content, placeholder, parts[1])
				}
			}
		}
		// Also replace {{.ENV_VAR}} patterns with environment variables
		content = replaceEnvVars(content)

		if sendTTL != "" {
			if _, err := utils.ParseDuration(sendTTL); err != nil {
				return fmt.Errorf("invalid --ttl: %w", err)
			}
		}

		if sendEncrypted {
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

		if sendID != "" && sendAlias != "" {
			return fmt.Errorf("--id and --alias are mutually exclusive")
		}

		var snippet *models.Snippet
		var err error

		if sendID != "" {
			// Resolve short prefix to full ID (like "597ebc3e" → full UUID)
			localDB, ldbErr := storage.NewLocalDB(localDBPath())
			if ldbErr != nil {
				return fmt.Errorf("open local db: %w", ldbErr)
			}
			resolver := resolve.NewResolver(localDB, client)
			resolved, _, resolveErr := resolver.GetByID(sendID)
			localDB.Close()
			if resolveErr != nil {
				return fmt.Errorf("resolve snippet id %q: %w", sendID, resolveErr)
			}
			snippet, err = client.UpdateSnippet(resolved.ID, content)
			if err != nil {
				return fmt.Errorf("update snippet: %w", err)
			}
			fmt.Printf("Snippet updated (id: %s) — version history saved\n", models.ShortID(snippet.ID))
		} else if sendAlias != "" {
			// Upsert by alias: update if exists, create if not
			existing, lookupErr := client.GetSnippetByAlias(sendAlias)
			if lookupErr == nil && existing != nil {
				snippet, err = client.UpdateSnippet(existing.ID, content)
				if err != nil {
					return fmt.Errorf("update snippet: %w", err)
				}
				fmt.Printf("Snippet updated (alias: %s, id: %s) — version history saved\n", snippet.Alias, snippet.ID)
			} else {
				snippet, err = client.CreateSnippetFull(content, sendAlias, sendDescription, sendTTL, sendEncrypted)
				if err != nil {
					return fmt.Errorf("send to cloud: %w", err)
				}
				fmt.Printf("Snippet sent to cloud with id: %s (alias: %s)\n", snippet.ID, snippet.Alias)
			}
		} else {
			// No --id, no --alias: always create new
			snippet, err = client.CreateSnippetFull(content, sendAlias, sendDescription, sendTTL, sendEncrypted)
			if err != nil {
				return fmt.Errorf("send to cloud: %w", err)
			}
			fmt.Printf("Snippet sent to cloud with id: %s\n", snippet.ID)
		}

		if snippet.ExpiresAt != nil {
			fmt.Printf("  [expires: %s]\n", snippet.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

// replaceEnvVars replaces {{.ENV_VAR}} patterns with environment variable values.
func replaceEnvVars(content string) string {
	re := regexp.MustCompile(`\{\{\.([A-Za-z_][A-Za-z0-9_]*)\}\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		varName := match[3 : len(match)-2] // strip {{. and }}
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return match // leave unchanged if env var not set
	})
}

func init() {
	sendCmd.Flags().StringVar(&sendID, "id", "", "update existing snippet by ID (creates version history)")
	sendCmd.Flags().StringVar(&sendAlias, "alias", "", "assign an alias name to this snippet (upsert: updates if exists)")
	sendCmd.Flags().StringVar(&sendDescription, "desc", "", "description for the snippet")
	sendCmd.Flags().StringVar(&sendTTL, "ttl", "", "time to live (e.g., 30s, 5m, 1h, 1d)")
	sendCmd.Flags().BoolVar(&sendEncrypted, "encrypt", false, "encrypt content with AES-256-GCM")
	sendCmd.Flags().StringSliceVar(&sendVars, "var", nil, "variable substitution: KEY=VALUE (replaces {{.KEY}})")
	rootCmd.AddCommand(sendCmd)
}
