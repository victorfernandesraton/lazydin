package config

import (
	"errors"

	"github.com/spf13/viper"
)

const (
	configUsername = "credentials.username"
	configPassword = "credentials.password"
)

// CredentialsConfig holds the credentials for the application
type CredentialsConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type Credentials struct {
	Username string
	Password string
}

func DefaultCredentials() {
	// Set default values for configuration options if necessary
	viper.SetDefault(configUsername, "user@mail.com")
	viper.SetDefault(configPassword, "user.pass")
}

func SetCredentials(username, password string) error {
	viper.Set(configUsername, username)
	viper.Set(configPassword, password)
	return viper.WriteConfig()
}

// LoadCredentials loads the Linkedin credentials from environment variables or flags
func LoadCredentials(config *Config, flagUsername, flagPassword string) (*Credentials, error) {
	envUsername := config.Credentials.Username
	envPassword := config.Credentials.Password

	if flagUsername != "" {
		envUsername = flagUsername
	}
	if flagPassword != "" {
		envPassword = flagPassword
	}
	credentials := &Credentials{Username: envUsername, Password: envPassword}

	if credentials.Username == "" || credentials.Password == "" {
		return nil, errors.New(
			"username and password must be set either via flags or environment variables",
		)
	}

	return credentials, nil
}
