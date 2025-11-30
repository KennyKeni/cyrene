package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port   int    `mapstructure:"PORT"`
	AppEnv string `mapstructure:"APP_ENV"`
	DB     DBConfig
}

type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	Database string `mapstructure:"DB_DATABASE"`
	Username string `mapstructure:"DB_USERNAME"`
	Password string `mapstructure:"DB_PASSWORD"`
	Schema   string `mapstructure:"DB_SCHEMA"`
}

var cfg Config

func Load() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Allow environment variables to override config file
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("APP_ENV", "local")
	viper.SetDefault("DB_SCHEMA", "public")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	cfg = Config{
		Port:   viper.GetInt("PORT"),
		AppEnv: viper.GetString("APP_ENV"),
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Database: viper.GetString("DB_DATABASE"),
			Username: viper.GetString("DB_USERNAME"),
			Password: viper.GetString("DB_PASSWORD"),
			Schema:   viper.GetString("DB_SCHEMA"),
		},
	}
}

func Get() *Config {
	return &cfg
}

func GetDB() *DBConfig {
	return &cfg.DB
}
