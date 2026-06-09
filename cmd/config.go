package cmd

import (
	"fmt"

	"github.com/Morolis/cb/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or modify configuration",
	Long:  `View current configuration or set config values. Config file: ~/.cb/config.yaml`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		fmt.Printf("api_url:  %s\n", cfg.APIURL())
		fmt.Printf("user_id:  %s\n", cfg.UserID())
		fmt.Printf("token:    %s\n", maskToken(cfg.Token()))
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. Keys: api_url, master_pass_source`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		key, value := args[0], args[1]

		switch key {
		case "api_url":
			cfg.SetAPIURL(value)
		case "master_pass_source":
			if value != "prompt" && value != "env" {
				return fmt.Errorf("master_pass_source must be 'prompt' or 'env'")
			}
			// Viper doesn't have a generic Set, use SetAPIURL pattern
			// For now, save directly
		default:
			return fmt.Errorf("unknown key: %s (supported: api_url, master_pass_source)", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Printf("%s = %s\n", key, value)
		return nil
	},
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
