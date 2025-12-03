package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv     string `mapstructure:"APP_ENV"`
	Server     ServerConfig
	DB         DBConfig
	Redis      RedisConfig
	Kafka      KafkaConfig
	Qdrant     QdrantConfig
	Genkit     GenkitConfig
	PokemonAPI PokemonAPIConfig
}

type ServerConfig struct {
	Addr string
}

type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	Database string `mapstructure:"DB_DATABASE"`
	Username string `mapstructure:"DB_USERNAME"`
	Password string `mapstructure:"DB_PASSWORD"`
	Schema   string `mapstructure:"DB_SCHEMA"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type KafkaConfig struct {
	Brokers       string `mapstructure:"KAFKA_BROKERS"`
	ConsumerGroup string `mapstructure:"KAFKA_CONSUMER_GROUP"`
}

type QdrantConfig struct {
	Host               string `mapstructure:"QDRANT_HOST"`
	Port               int    `mapstructure:"QDRANT_PORT"`
	APIKey             string `mapstructure:"QDRANT_API_KEY"`
	Collection         string `mapstructure:"QDRANT_COLLECTION"`
	CollectionDim      uint   `mapstructure:"QDRANT_COLLECTION_DIM"`
	CacheCollection    string `mapstructure:"QDRANT_CACHE_COLLECTION"`
	CacheCollectionDim uint   `mapstructure:"QDRANT_CACHE_COLLECTION_DIM"`
}

type GenkitConfig struct {
	EmbedURL    string `mapstructure:"EMBED_URL"`
	EmbedAPIKey string `mapstructure:"EMBED_API_KEY"`
	EmbedModel  string `mapstructure:"EMBED_MODEL"`
	AgentURL    string `mapstructure:"AGENT_URL"`
	AgentAPIKey string `mapstructure:"AGENT_API_KEY"`
	AgentModel  string `mapstructure:"AGENT_MODEL"`
}

type PokemonAPIConfig struct {
	BaseURL string `mapstructure:"POKEMON_BASE_URL"`
	//APIKey  string `mapstructure:"POKEMON_API_KEY"`
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
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_CONSUMER_GROUP", "cyrene")
	viper.SetDefault("QDRANT_HOST", "localhost")
	viper.SetDefault("QDRANT_PORT", "6334")
	viper.SetDefault("QDRANT_API_KEY", "")
	viper.SetDefault("QDRANT_COLLECTION", "cobblemon")
	viper.SetDefault("QDRANT_COLLECTION_DIM", 4096)
	viper.SetDefault("QDRANT_CACHE_COLLECTION", "cache")
	viper.SetDefault("QDRANT_CACHE_COLLECTION_DIM", 1024)
	viper.SetDefault("EMBED_URL", "https://openrouter.ai/api/v1/embeddings")
	viper.SetDefault("AGENT_URL", "https://openrouter.ai/api/v1")
	viper.SetDefault("EMBED_MODEL", "qwen/qwen3-embedding-8b")
	viper.SetDefault("AGENT_MODEL", "openai/gpt-oss-120b")
	//viper.SetDefault("POKEMON_API_KEY", "")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			log.Printf("Error reading config file: %v", err)
		}
	}

	cfg = Config{
		AppEnv: viper.GetString("APP_ENV"),
		Server: ServerConfig{
			Addr: fmt.Sprintf(":%d", viper.GetInt("PORT")),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Database: viper.GetString("DB_DATABASE"),
			Username: viper.GetString("DB_USERNAME"),
			Password: viper.GetString("DB_PASSWORD"),
			Schema:   viper.GetString("DB_SCHEMA"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Kafka: KafkaConfig{
			Brokers:       viper.GetString("KAFKA_BROKERS"),
			ConsumerGroup: viper.GetString("KAFKA_CONSUMER_GROUP"),
		},
		Qdrant: QdrantConfig{
			Host:   viper.GetString("QDRANT_HOST"),
			Port:   viper.GetInt("QDRANT_PORT"),
			APIKey: viper.GetString("QDRANT_API_KEY"),
		},
		Genkit: GenkitConfig{
			EmbedURL:    viper.GetString("EMBED_URL"),
			EmbedAPIKey: viper.GetString("EMBED_API_KEY"),
			EmbedModel:  viper.GetString("EMBED_MODEL"),
			AgentURL:    viper.GetString("AGENT_URL"),
			AgentAPIKey: viper.GetString("AGENT_API_KEY"),
			AgentModel:  viper.GetString("AGENT_MODEL"),
		},
		PokemonAPI: PokemonAPIConfig{
			BaseURL: viper.GetString("POKEMON_BASE_URL"),
		},
	}
}

func Get() *Config {
	return &cfg
}

func GetDB() *DBConfig {
	return &cfg.DB
}

func GetRedis() *RedisConfig {
	return &cfg.Redis
}

func GetKafka() *KafkaConfig {
	return &cfg.Kafka
}

func GetQdrant() *QdrantConfig {
	return &cfg.Qdrant
}

func GetGenkit() *GenkitConfig { return &cfg.Genkit }
