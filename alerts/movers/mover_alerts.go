package movers

import (
	"context"
	"github.com/DrGrimshaw/gohtml"
	"stocks/alerts"
	"stocks/alerts/movers/morning_star"
	"stocks/models"
)

type Config struct {
	MSAPI morning_star.MSAPI
}
type implementer struct {
	config Config
}

func (i *implementer) GetAlerts(ctx context.Context, holdingsMap map[string]models.Holding) ([]alerts.Alert, []alerts.Subscriber, error) {
	movers, err := i.config.MSAPI.GetMovers(ctx)
	if err != nil {
		return nil, nil, err
	}

	var retAlerts []string
	var checkers = []struct {
		holdings []models.MSHolding
		action   string
	}{
		{holdings: movers.Actives, action: "Active"},
		{holdings: movers.Losers, action: "Loser"},
		{holdings: movers.Gainers, action: "Gainer"},
	}
	for _, checker := range checkers {
		alert, err := retrieveHTMLAlert(checker.holdings, holdingsMap, checker.action)
		if err != nil {
			return nil, nil, err
		}
		retAlerts = append(retAlerts, alert)
	}
	subscribersFromYaml, err := alerts.LoadSubscribersFromYaml("./alerts/movers/subscribers.yaml")
	if err != nil {
		return nil, nil, err
	}
	return retAlerts, subscribersFromYaml, err
}

type AlertHTML struct {
	Action           string  `html:"l=Action,e=span,c=action"`
	Ticker           string  `html:"l=Ticker,e=span,c=ticker"`
	Name             string  `html:"l=Name,e=span,c=name"`
	PercentChange    float64 `html:"l=Percent Change,e=span,c=percent_change"`
	LastPrice        float64 `html:"l=Last Price,e=span,c=last_price"`
	Nothing          string  `html:"l=|,e=span,c=action"`
	LeveragedETF     string  `html:"l=Leveraged ETF,e=span,c=leveraged_etf"`
	PercentOwnership float64 `html:"l=Percentage ownership in leveraged etf,e=span,c=percent_ownership"`
}

type AlertHTMLArr struct {
	Alerts []AlertHTML `html:"row"`
}

func retrieveHTMLAlert(movers []models.MSHolding, holdingsMap map[string]models.Holding, action string) (string, error) {
	alertHTMLS, err := retrieveAlerts(movers, holdingsMap, action)
	if err != nil {
		return "", err
	}
	return gohtml.Encode(AlertHTMLArr{Alerts: alertHTMLS})
}

func retrieveAlerts(movers []models.MSHolding, holdingsMap map[string]models.Holding, action string) ([]AlertHTML, error) {
	var retAlerts []AlertHTML
	for _, mover := range movers {
		if holding, found := holdingsMap[mover.Ticker]; found {
			alertHTML := AlertHTML{
				Action:           action,
				Ticker:           mover.Ticker,
				Name:             mover.Name,
				PercentChange:    mover.PercentNetChange,
				LastPrice:        mover.LastPrice,
				Nothing:          " ",
				LeveragedETF:     holding.AccountTicker,
				PercentOwnership: holding.Percent,
			}
			//htmlEncode := fmt.Sprintf("found %s stock ticker %+v in holding %+v\n", action, mover, holding)
			retAlerts = append(retAlerts, alertHTML)
		}
	}
	return retAlerts, nil
}

func New(config Config) alerts.AlertParser {
	return &implementer{config: config}
}
