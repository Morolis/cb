package cmd

import (
	"fmt"

	"github.com/Morolis/cb/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove local authentication token",
	Long:  `Remove the locally stored API token and user ID from your configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		cfg.ClearToken()
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Println("Logged out successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
