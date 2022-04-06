package main

import (
	"context"
	"fmt"
	"sort"
	"stocks/database"
	"stocks/direxion"
	"stocks/models"
	"stocks/morning_star"
)

func orchestrate(
	ctx context.Context,
	db database.DB,
	client direxion.Client,
	msapi morning_star.MSAPI) error {
	seeds, err := db.ListSeeds(ctx)
	if err != nil {
		return err
	}

	holdings, err := fetchHoldings(ctx, seeds, client)
	if err != nil {
		return err
	}

	holdingsMap := make(map[string]models.Holding)
	for _, holding := range holdings {
		holdingsMap[holding.StockTicker] = holding
	}
	fmt.Println(holdingsMap)

	movers, err := msapi.GetMovers(ctx)
	if err != nil {
		return err
	}
	fmt.Println(movers)
	var alerts []string
	alerts = append(alerts, retrieveAlerts(movers.Actives, holdingsMap, "active")...)
	alerts = append(alerts, retrieveAlerts(movers.Losers, holdingsMap, "loser")...)
	alerts = append(alerts, retrieveAlerts(movers.Gainers, holdingsMap, "gainer")...)
	fmt.Printf("Found alerts: %s\n", alerts)
	return nil
}

func retrieveAlerts(movers []models.MSHolding, holdingsMap map[string]models.Holding, action string) []string {
	var alerts []string
	for _, mover := range movers {
		if holding, found := holdingsMap[mover.Ticker]; found {
			alerts = append(alerts, fmt.Sprintf("found %s stock ticker %+v in holding %+v\n", action, mover, holding))
		}
	}
	return alerts
}

func fetchHoldings(ctx context.Context, seeds []models.Seed, client direxion.Client) ([]models.Holding, error) {
	var allHoldings []models.Holding
	for _, seed := range seeds {
		fmt.Printf("fetching information for %+v\n", seed)
		holdings, err := client.GetHoldings(ctx, seed)
		if err != nil {
			return nil, err
		}
		allHoldings = append(allHoldings, holdings...)
	}

	sort.Slice(allHoldings, func(i, j int) bool {
		return allHoldings[i].Percent > allHoldings[j].Percent
	})

	return allHoldings, nil
}

func main() {
	ctx := context.Background()
	db := database.NewDumbDatabase()
	client, err := direxion.NewDirexionClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	msapi := morning_star.New(morning_star.Config{
		URL:  "https://ms-finance.p.rapidapi.com/market/v2/get-movers",
		Host: "ms-finance.p.rapidapi.com",
		Key:  "e6e18b1891mshe45bf4b11c2c441p199735jsn2958e367084e",
	})

	err = orchestrate(ctx, db, client, msapi)
	if err != nil {
		fmt.Println(err)
		return
	}
}
