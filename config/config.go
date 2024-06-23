package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config struct holds the configuration options for the application
type Config struct {
	Credentials CredentialsConfig `mapstructure:"credentials"`
	SQlite      string            `mapstructure:"storage"`
}

// LoadConfig loads the configuration from file or environment variables
func LoadConfig() (*Config, error) {

	home, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf(err.Error())

	}
	appPath := filepath.Join(home, "lazydin")

	configPath := filepath.Join(appPath, "config.toml")
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	DefaultCredentials()
	DefaultStorage(appPath)

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
