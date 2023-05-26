package initializers

import (
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost         string `mapstructure:"POSTGRES_HOST"`     // POSTGRES_HOST
	DBUserName     string `mapstructure:"POSTGRES_USER"`     // POSTGRES_USER
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"` // POSTGRES_PASSWORD
	DBName         string `mapstructure:"POSTGRES_DB"`       // POSTGRES_DB
	DBPort         string `mapstructure:"POSTGRES_PORT"`     // POSTGRES_PORT

	JwtSecret    string        `mapstructure:"JWT_SECRET"`     // JWT_SECRET
	JwtExpiresIn time.Duration `mapstructure:"JWT_EXPIRED_IN"` // JWT_EXPIRED_IN
	JwtMaxAge    int           `mapstructure:"JWT_MAXAGE"`     // JWT_MAXAGE

	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"` // CLIENT_ORIGIN
}

var (
	onceEnv sync.Once // guards Env initialization
	config  Config    // global connection to the Env
)

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (cfg *Config, err error) {
	onceEnv.Do(func() {
		viper.AddConfigPath(path)
		viper.SetConfigType("env")
		viper.SetConfigName(".env")
		viper.AutomaticEnv()

		err = viper.ReadInConfig()
		if err != nil {
			return
		}
		err = viper.Unmarshal(&config)
	})
	return &config, err
}
