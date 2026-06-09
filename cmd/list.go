package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/Morolis/cb/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	listLimit  int
	listOffset int
	listSource string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your snippets (local + cloud)",
	Long:  `Display a table of all saved snippets with source indicator, alias, preview, and timestamps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)
		localDB, err := storage.NewLocalDB(localDBPath())
		if err != nil {
			return fmt.Errorf("open local db: %w", err)
		}
		defer localDB.Close()

		resolver := resolve.NewResolver(localDB, client)

		previews, sources, err := resolver.ListMerged(listLimit, listSource)
		if err != nil {
			return fmt.Errorf("list snippets: %w", err)
		}

		if len(previews) == 0 {
			fmt.Println("No snippets found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SOURCE\tALIAS\tDESC\tID\tPREVIEW\tCREATED\tEXPIRES")
		fmt.Fprintln(w, "------\t-----\t----\t--\t-------\t-------\t-------")

		for i, s := range previews {
			alias := s.Alias
			if alias == "" {
				alias = "-"
			}
			desc := s.Description
			if desc == "" {
				desc = "-"
			} else {
				desc = utils.TruncateString(desc, 20)
			}
			preview := s.Preview
			if s.Encrypted {
				preview = "[encrypted]"
			} else {
				preview = models.SanitizePreview(preview)
				preview = utils.TruncateString(preview, 50)
			}
			created := s.CreatedAt.Format("2006-01-02 15:04")
			expires := "-"
			if s.ExpiresAt != nil {
				expires = s.ExpiresAt.Format("2006-01-02 15:04")
			}
			shortID := models.ShortID(s.ID)
			source := string(sources[i])
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", source, alias, desc, shortID, preview, created, expires)
		}

		w.Flush()
		return nil
	},
}

func init() {
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "number of snippets to display")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "offset for pagination")
	listCmd.Flags().StringVar(&listSource, "source", "all", "filter by source: all, local, remote")
	rootCmd.AddCommand(listCmd)
}
