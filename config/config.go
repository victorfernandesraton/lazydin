package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config struct holds the configuration options for the application
type Config struct {
	Credentials CredentialsConfig `mapstructure:"credentials"`
}

// LoadConfig loads the configuration from file or environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	DefaultCredentials()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, err
		}
		if err := viper.WriteConfigAs(configPath); err != nil {
			return nil, err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
