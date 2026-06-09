package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginUser string
	loginPass string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to cb and save API token",
	Long:  `Authenticate with the cb server. If no account exists, you will be registered automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		client := api.NewClient(cfg)

		username := loginUser
		if username == "" {
			fmt.Print("Username: ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			username = strings.TrimSpace(line)
		}

		password := loginPass
		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			fmt.Println()
			password = string(bytePassword)
		}

		// Try login first
		resp, err := client.Login(username, password)
		if err != nil {
			// Only try registration if user doesn't exist
			if strings.Contains(err.Error(), "user not found") {
				if len(password) < 6 {
					return fmt.Errorf("password must be at least 6 characters")
				}
				fmt.Println("User not found, creating account...")
				resp, err = client.Register(username, password)
				if err != nil {
					return fmt.Errorf("registration failed: %w", err)
				}
				fmt.Println("Account created successfully!")
			} else {
				return fmt.Errorf("login failed: %w", err)
			}
		}

		if err := cfg.SaveToken(resp.Token); err != nil {
			return fmt.Errorf("save token: %w", err)
		}
		if err := cfg.SaveUserID(resp.UserID); err != nil {
			return fmt.Errorf("save user id: %w", err)
		}

		// Save api-url to config if explicitly provided
		if apiURL != "" {
			cfg.SetAPIURL(apiURL)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save api url: %w", err)
			}
		}

		fmt.Printf("Logged in as %s!\n", resp.Username)
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginUser, "user", "", "username")
	loginCmd.Flags().StringVar(&loginPass, "password", "", "password (not recommended, use interactive prompt)")
	rootCmd.AddCommand(loginCmd)
}
