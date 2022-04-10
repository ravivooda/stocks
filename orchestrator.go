package main

import (
	"context"
	"fmt"
	"sort"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/models"
	"stocks/securities"
	"stocks/securities/direxion"
)

func orchestrate(ctx context.Context, db database.DB, client securities.Client, parsers []alerts.AlertParser) error {
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

	var gatheredAlerts []alerts.Alert
	for _, parser := range parsers {
		tAlerts, _, err := parser.GetAlerts(ctx, holdingsMap)
		if err != nil {
			return err
		}
		gatheredAlerts = append(gatheredAlerts, tAlerts...)
	}
	fmt.Printf("Found alerts: %s\n", gatheredAlerts)
	return nil
}

func fetchHoldings(ctx context.Context, seeds []models.Seed, client securities.Client) ([]models.Holding, error) {
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

	config, err := NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(config)

	msapi := morning_star.New(config.MSAPI)

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	err = orchestrate(ctx, db, client, alertParsers)
	if err != nil {
		fmt.Println(err)
		return
	}
}
