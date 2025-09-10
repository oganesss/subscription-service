package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HTTP struct {
		Address string `mapstructure:"address"`
		ReadTimeoutSeconds int `mapstructure:"read_timeout_seconds"`
		WriteTimeoutSeconds int `mapstructure:"write_timeout_seconds"`
		IdleTimeoutSeconds int `mapstructure:"idle_timeout_seconds"`
	} `mapstructure:"http"`

	Postgres struct {
		DSN string `mapstructure:"dsn"`
		MaxConns int32 `mapstructure:"max_conns"`
		MinConns int32 `mapstructure:"min_conns"`
	} `mapstructure:"postgres"`

	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	v.SetEnvPrefix("SUBS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("http.address", ":8080")
	v.SetDefault("http.read_timeout_seconds", 10)
	v.SetDefault("http.write_timeout_seconds", 10)
	v.SetDefault("http.idle_timeout_seconds", 60)
	v.SetDefault("postgres.max_conns", 10)
	v.SetDefault("postgres.min_conns", 2)
	v.SetDefault("log.level", "info")

	_ = v.ReadInConfig()

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &c, nil
}



