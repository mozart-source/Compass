package config

import (
	"github.com/spf13/viper"
)



func LoadDatabaseConfig() *DatabaseConfig {
    return &DatabaseConfig{
        Host:     viper.GetString("DB_HOST"),
        Port:     viper.GetInt("DB_PORT"),
        User:     viper.GetString("DB_USER"),
        Password: viper.GetString("DB_PASSWORD"),
        Name:   viper.GetString("DB_NAME"),
        SSLMode:  viper.GetString("DB_SSLMODE"),
    }
}