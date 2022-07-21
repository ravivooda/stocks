package main

import (
	"github.com/spf13/viper"
	"stocks/alerts/movers/morning_star"
	"stocks/insights/overlap"
	"stocks/securities/masterdatareports"
	"stocks/securities/proshares"
)

type Secrets struct {
	MSAPI struct {
		Key string
	} `mapstructure:"ms_api"`
	Notifications struct {
		ShouldSendEmails bool `mapstructure:"should_send_email"`
	} `mapstructure:"notifications"`
	Uploads struct {
		ShouldUploadInsightsOutputToGCP bool `mapstructure:"should_upload_insights_output_to_gcp"`
	} `mapstructure:"uploads"`
}

type Config struct {
	MSAPI       morning_star.Config `mapstructure:"ms_api"`
	Directories struct {
		Temporary string `mapstructure:"tmp"`
		Build     string `mapstructure:"build"`
		Artifacts string `mapstrucutre:"artifacts"`
		Websites  string `mapstructure:"websites"`
	} `mapstructure:"directories"`
	Securities struct {
		ProShares         proshares.Config         `mapstructure:"pro_shares"`
		MasterDataReports masterdatareports.Config `mapstructure:"master_data_reports"`
	} `mapstructure:"securities"`
	Secrets Secrets
	Outputs struct {
		Insights overlap.Config
		Websites struct {
			MinThresholdPercentage int `mapstructure:"min_threshold_percentage"`
		}
	}
}

func NewConfig() (Config, error) {
	v, err := loadViperConfig("./config.yaml")
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = v.Unmarshal(&c)
	if err != nil {
		return Config{}, err
	}

	s, err := loadViperConfig("./secrets.yaml")
	if err != nil {
		return Config{}, err
	}
	var se Secrets
	err = s.Unmarshal(&se)
	if err != nil {
		return Config{}, err
	}

	c.Secrets = se
	c.MSAPI.Key = c.Secrets.MSAPI.Key

	return c, nil
}

func loadViperConfig(filepath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filepath)
	err := v.ReadInConfig()
	return v, err
}
