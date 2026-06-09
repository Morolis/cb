package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Morolis/cb/internal/api"
	"github.com/Morolis/cb/internal/config"
	"github.com/Morolis/cb/internal/crypto"
	"github.com/Morolis/cb/internal/models"
	"github.com/Morolis/cb/internal/resolve"
	"github.com/Morolis/cb/internal/storage"
	"github.com/spf13/cobra"
)

var (
	execYes  bool
	execArgs []string
)

var execCmd = &cobra.Command{
	Use:   "exec <alias|id>",
	Short: "Execute a snippet as a shell command",
	Long:  `Execute a saved snippet in the local shell. Works with both local and remote snippets.`,
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
		var snippet *models.Snippet

		if models.IsLocalID(id) {
			snippet, _, err = resolver.GetByID(id)
		} else {
			snippet, _, err = resolver.GetByAlias(id)
			if err != nil {
				snippet, _, err = resolver.GetByID(id)
			}
		}
		if err != nil {
			return fmt.Errorf("get snippet: %w", err)
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

		command := snippet.Content
		if len(execArgs) > 0 {
			command = command + " " + strings.Join(execArgs, " ")
		}

		if !execYes {
			fmt.Printf("About to execute:\n  %s\n", command)
			fmt.Print("Proceed? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		var shell, flag string
		if runtime.GOOS == "windows" {
			shell, flag = "cmd", "/C"
		} else {
			shell, flag = "sh", "-c"
		}

		runCmd := exec.Command(shell, flag, command)
		runCmd.Stdin = os.Stdin
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr

		if err := runCmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			return fmt.Errorf("execute: %w", err)
		}

		return nil
	},
}

func init() {
	execCmd.Flags().BoolVar(&execYes, "yes", false, "skip confirmation prompt")
	execCmd.Flags().StringSliceVar(&execArgs, "args", nil, "additional arguments to append to the command")
	rootCmd.AddCommand(execCmd)
}
