package movers

import (
	"context"
	"fmt"
	"stocks/alerts"
	"stocks/models"
	"stocks/morning_star"
)

type Config struct {
	MSAPI morning_star.MSAPI
}
type implementer struct {
	config Config
}

func (i *implementer) GetAlerts(ctx context.Context, holdingsMap map[string]models.Holding) ([]alerts.Alert, error) {
	movers, err := i.config.MSAPI.GetMovers(ctx)
	if err != nil {
		return nil, err
	}

	var retAlerts []string
	retAlerts = append(retAlerts, retrieveAlerts(movers.Actives, holdingsMap, "active")...)
	retAlerts = append(retAlerts, retrieveAlerts(movers.Losers, holdingsMap, "loser")...)
	retAlerts = append(retAlerts, retrieveAlerts(movers.Gainers, holdingsMap, "gainer")...)
	return retAlerts, err
}

func retrieveAlerts(movers []models.MSHolding, holdingsMap map[string]models.Holding, action string) []string {
	var retAlerts []string
	for _, mover := range movers {
		if holding, found := holdingsMap[mover.Ticker]; found {
			retAlerts = append(retAlerts, fmt.Sprintf("found %s stock ticker %+v in holding %+v\n", action, mover, holding))
		}
	}
	return retAlerts
}

func New(config Config) alerts.AlertParser {
	return &implementer{config: config}
}
