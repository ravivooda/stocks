package orchestrate

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"stocks/alerts/movers/morning_star"
	"stocks/insights/overlap"
	"stocks/securities/masterdatareports"
	"stocks/securities/proshares"
	"stocks/utils"
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
	TestConfig struct {
		MaxServerRunTime int `mapstructure:"max_server_run_time"`
	} `mapstructure:"test_config"`
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
	content, _ := ioutil.ReadFile("config.yaml")
	utils.PanicErrWithExtraMessage(err, string(content))
	var c Config
	utils.PanicErr(v.Unmarshal(&c))

	s, err := loadViperConfig("./secrets.yaml")
	content, _ = ioutil.ReadFile("secrets.yaml")
	utils.PanicErrWithExtraMessage(err, string(content))

	var se Secrets
	utils.PanicErr(s.Unmarshal(&se))

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
