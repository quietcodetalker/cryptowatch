package config

import "github.com/spf13/viper"

// Config represents whole app configuration.
type Config struct {
	SymmetricKey       string `mapstructure:"SYMMETRIC_KEY" validate:"required"`
	BindAddr           string `mapstructure:"BIND_ADDR" validate:"required"`
	DBHost             string `mapstructure:"DB_HOST" validate:"required"`
	DBPort             string `mapstructure:"DB_PORT" validate:"required"`
	DBUser             string `mapstructure:"DB_USER" validate:"required"`
	DBPassword         string `mapstructure:"DB_PASSWORD" validate:"required"`
	DBName             string `mapstructure:"DB_NAME" validate:"required"`
	DBSSLMode          string `mapstructure:"DB_SSLMODE" validate:"required"`
	TelegramToken      string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`
	CryptoCompareToken string `mapstructure:"CRYPTOCOMPARE_TOKEN" validate:"required"`
}

func LoadConfig(path string, name string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
