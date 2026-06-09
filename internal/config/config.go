package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	v *viper.Viper
}

func Get() *Config {
	once.Do(func() {
		instance = &Config{v: viper.New()}
	})
	return instance
}

func (c *Config) SetConfigFile(path string) {
	c.v.SetConfigFile(path)
}

func (c *Config) SetAPIURL(url string) {
	c.v.Set("api_url", url)
}

func (c *Config) Load() error {
	dir, err := cbDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	c.v.SetConfigName("config")
	c.v.SetConfigType("yaml")
	c.v.AddConfigPath(dir)

	c.v.SetDefault("api_url", "http://localhost:8080/v1")
	c.v.SetDefault("master_pass_source", "prompt")

	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		// Config file not found — write defaults
		cfgPath := filepath.Join(dir, "config.yaml")
		if err := c.v.WriteConfigAs(cfgPath); err != nil {
			return fmt.Errorf("write default config: %w", err)
		}
	}
	return nil
}

func (c *Config) APIURL() string {
	return c.v.GetString("api_url")
}

func (c *Config) Token() string {
	return c.v.GetString("token")
}

func (c *Config) SetToken(token string) {
	c.v.Set("token", token)
}

func (c *Config) UserID() string {
	return c.v.GetString("user_id")
}

func (c *Config) SetUserID(id string) {
	c.v.Set("user_id", id)
}

func (c *Config) MasterPassSource() string {
	return c.v.GetString("master_pass_source")
}

func (c *Config) Save() error {
	dir, err := cbDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config.yaml")
	return c.v.WriteConfigAs(cfgPath)
}

func (c *Config) SaveToken(token string) error {
	c.SetToken(token)
	return c.Save()
}

func (c *Config) ClearToken() {
	c.v.Set("token", "")
	c.v.Set("user_id", "")
}

func (c *Config) SaveUserID(id string) error {
	c.SetUserID(id)
	return c.Save()
}

func CBDir() (string, error) {
	return cbDir()
}

func cbDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cb"), nil
}

func TokenFilePath() (string, error) {
	dir, err := cbDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token"), nil
}
