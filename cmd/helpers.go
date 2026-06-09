package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Morolis/cb/internal/config"
	"golang.org/x/term"
)

func localDBPath() string {
	dir, _ := config.CBDir()
	return filepath.Join(dir, "local.db")
}

func getMasterPassword() (string, error) {
	cfg := config.Get()
	source := cfg.MasterPassSource()

	// Always check env var first as a convenience
	if source == "env" || os.Getenv("CB_MASTER_PASS") != "" {
		pass := os.Getenv("CB_MASTER_PASS")
		if pass == "" {
			return "", fmt.Errorf("CB_MASTER_PASS environment variable not set")
		}
		return pass, nil
	}

	fmt.Print("Master password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	fmt.Println()
	return string(bytePassword), nil
}
