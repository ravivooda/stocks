package main

import (
	"github.com/spf13/viper"
	"stocks/alerts/movers/morning_star"
)

type Config struct {
	MSAPI         morning_star.Config `mapstructure:"ms_api"`
	Notifications struct {
		ShouldSendEmails bool `mapstructure:"should_send_email"`
	} `mapstructure:"notifications"`
	Uploads struct {
		ShouldUploadInsightsOutputToGCP bool `mapstructure:"should_upload_insights_output_to_gcp"`
	} `mapstructure:"uploads"`
	Directories struct {
		Temporary string `mapstructure:"tmp"`
		Build     string `mapstructure:"build"`
		Artifacts string `mapstrucutre:"artifacts"`
		Websites  string `mapstructure:"websites"`
	} `mapstructure:"directories"`
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
