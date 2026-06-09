package cmd

import (
	"fmt"
	"os"

	"github.com/Morolis/cb/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	apiURL  string
	verbose bool
	version string
	commit  string
	date    string
)

var rootCmd = &cobra.Command{
	Use:   "cb",
	Short: "Cross-device clipboard & code snippet sync tool",
	Long: `cb is a lightweight CLI tool for developers to sync short text,
code snippets, and commands across devices with end-to-end encryption support.`,
	SilenceUsage: true,
}

func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", v, c, d)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.cb/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "override server API URL")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug output")
}

func initConfig() {
	cfg := config.Get()
	if cfgFile != "" {
		cfg.SetConfigFile(cfgFile)
	}
	if apiURL != "" {
		cfg.SetAPIURL(apiURL)
	}
	if err := cfg.Load(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "config warning: %v\n", err)
		}
	}
}
