package config

import (
	"os"
	"path"

	"github.com/spf13/viper"
)

const (
	configSqlite = "storage"
)

type StorageConfig struct {
	Sqlite string
}

func DefaultStorage(configPath string) {
	sqlitePath := path.Join(configPath, "database.sqlite3")
	viper.SetDefault(configSqlite, sqlitePath)
}

func SetStorage(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Create(filePath); err != nil {
			return err
		}
	}
	viper.Set(configSqlite, filePath)
	return viper.WriteConfig()
}
