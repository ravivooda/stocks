package alerts

import (
	"context"
	"github.com/spf13/viper"
	"stocks/models"
)

type Alert = string

type Subscriber struct {
	Name  string
	Email string
}

type AlertParser interface {
	GetAlerts(ctx context.Context, holdingsMap map[models.StockTicker]models.LETFHolding) ([]Alert, []Subscriber, error)
}

func LoadSubscribersFromYaml(filename string) ([]Subscriber, error) {
	type Config struct {
		Subscribers []Subscriber
	}
	v := viper.New()
	v.SetConfigFile(filename)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var c Config
	err = v.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	return c.Subscribers, nil
}
