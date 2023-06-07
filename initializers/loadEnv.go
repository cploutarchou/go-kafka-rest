package initializers

import (
	"log"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	config      *Config
	configMutex sync.Mutex
)

type Config struct {
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUserName     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_DB"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`

	JwtSecret    string        `mapstructure:"JWT_SECRET"`
	JwtExpiresIn time.Duration `mapstructure:"JWT_EXPIRED_IN"`
	JwtMaxAge    int           `mapstructure:"JWT_MAXAGE"`

	ClientOrigin       string   `mapstructure:"CLIENT_ORIGIN"`
	CorsAllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`

	KafkaBrokers         string `mapstructure:"KAFKA_BROKERS"`
	KafkaNumOfPartitions int    `mapstructure:"KAFKA_NUM_OF_PARTITIONS"`

	EnableWebsocket bool `mapstructure:"ENABLE_WEBSOCKET"`
}

func LoadConfig(path string) (*Config, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	err := readAndParseConfig(path)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func readAndParseConfig(path string) error {
	setupViper(path)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return unmarshalConfig()
}

func setupViper(path string) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("sample")
	viper.AutomaticEnv()
}

func unmarshalConfig() error {
	config = &Config{}
	err := viper.Unmarshal(config)
	if err != nil {
		return err
	}
	return nil
}

func GetConfig() *Config {
	if config == nil {
		//reload config

		loadConfig, err := LoadConfig("./..")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		config = loadConfig
	}
	return config
}
