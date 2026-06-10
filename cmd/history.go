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

var historyCmd = &cobra.Command{
	Use:   "history [id|alias]",
	Short: "List version history of a snippet",
	Long:  `Display all past versions of a snippet. Only works for remote (cloud) snippets.`,
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
		snippet, source, err := resolver.GetByAlias(id)
		if err != nil {
			snippet, source, err = resolver.GetByID(id)
		}
		if err != nil {
			return fmt.Errorf("resolve snippet: %w", err)
		}

		if source == "local" {
			return fmt.Errorf("version history is only available for remote snippets")
		}

		versions, err := client.ListVersions(snippet.ID)
		if err != nil {
			return fmt.Errorf("list versions: %w", err)
		}

		if len(versions) == 0 {
			fmt.Println("No version history found.")
			return nil
		}

		alias := snippet.Alias
		if alias == "" {
			alias = snippet.ID[:8]
		}
		fmt.Printf("Version history for: %s\n\n", alias)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tCONTENT\tDATE")
		fmt.Fprintln(w, "-------\t-------\t----")

		for _, v := range versions {
			preview := models.SanitizePreview(v.Content)
			preview = utils.TruncateString(preview, 60)
			date := v.CreatedAt.Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%d\t%s\t%s\n", v.ID, preview, date)
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
