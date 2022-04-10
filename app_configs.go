package main

import (
	"github.com/spf13/viper"
	"stocks/alerts/movers/morning_star"
)

type Config struct {
	MSAPI morning_star.Config `mapstructure:"ms_api"`
}

func NewConfig() (Config, error) {
	v := viper.New()
	v.SetConfigFile("./config.yaml")
	err := v.ReadInConfig()
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = v.Unmarshal(&c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
