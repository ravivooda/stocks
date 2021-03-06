package database

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"stocks/models"
)

type dumbDatabase struct {
}

func (i dumbDatabase) ListSeeds(_ context.Context) ([]models.Seed, error) {
	type YAML struct {
		Seeds struct {
			Direxion struct {
				UrlBase string `mapstructure:"url_base"`
				Simple  struct {
					Tickers []string
					Header  models.Header
				}
				Complicated struct {
					Tickers []string
					Header  models.Header
				}
			}
			MicroSector struct {
				Simple struct {
					Tickers []string
				}
			} `mapstructure:"microsector"`
		}
	}

	v := viper.New()
	v.SetConfigFile("./database/seeds.yaml")
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var c YAML
	err = v.Unmarshal(&c)
	if err != nil {
		return nil, err
	}
	fmt.Println(c)
	var rets []models.Seed
	for _, stock := range c.Seeds.Direxion.Simple.Tickers {
		rets = append(rets, models.Seed{
			URL:      fmt.Sprintf("%s/%s.csv", c.Seeds.Direxion.UrlBase, stock),
			Ticker:   stock,
			Header:   c.Seeds.Direxion.Simple.Header,
			Provider: models.Direxion,
		})
	}

	for _, stock := range c.Seeds.Direxion.Complicated.Tickers {
		rets = append(rets, models.Seed{
			URL:      fmt.Sprintf("%s/%s.csv", c.Seeds.Direxion.UrlBase, stock),
			Ticker:   stock,
			Header:   c.Seeds.Direxion.Complicated.Header,
			Provider: models.Direxion,
		})
	}

	for _, stock := range c.Seeds.MicroSector.Simple.Tickers {
		rets = append(rets, models.Seed{
			URL:      "",
			Ticker:   stock,
			Header:   models.Header{},
			Provider: models.MicroSector,
		})
	}

	return rets, nil
}

func NewDumbDatabase() DB {
	return &dumbDatabase{}
}
