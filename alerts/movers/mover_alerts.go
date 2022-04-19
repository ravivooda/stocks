package movers

import (
	"context"
	"github.com/DrGrimshaw/gohtml"
	"stocks/alerts"
	"stocks/alerts/movers/morning_star"
	"stocks/models"
	"strings"
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
	Nothing          string  `html:"l=,e=span,c=action"`
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
	encodedHTML, err := gohtml.Encode(AlertHTMLArr{Alerts: alertHTMLS})
	if err != nil {
		return "", err
	}
	return appendCSS(encodedHTML), nil
}

type validator func(int) bool

func appendCSS(encodedHTML string) string {
	replaceConfigs := []struct {
		old       string
		new       string
		validator validator
	}{
		{old: "<td>", new: "<td style=\"border: 1px solid #ddd; padding: 8px;\">"},
		{old: "<tr>", new: "<tr style=\"background-color: #f2f2f2;\">", validator: func(i int) bool {
			return i%2 == 0
		}},
		{old: "<thead>", new: "<thead style=\"padding-top: 12px;padding-bottom: 12px;text-align: left;background-color: #04AA6D;color: white;\">"},
		{old: "<table>", new: "<table style=\"padding-bottom: 50px\">"},
	}
	var retString = encodedHTML
	for _, config := range replaceConfigs {
		var i = 1
		placeholder := "DOESNT_MATTER_WHAT_WE_PUT_HERE_ITS_JUST_A_SIMPLE_PLACEHOLDER"
		for true {
			if config.validator == nil || config.validator(i) {
				retString = strings.Replace(retString, config.old, config.new, 1)
			} else {
				retString = strings.Replace(retString, config.old, placeholder, 1)
			}
			i += 1
			if !strings.Contains(retString, config.old) {
				break
			}
		}
		retString = strings.ReplaceAll(retString, placeholder, config.old)
	}
	return retString
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
