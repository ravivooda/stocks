package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/models"
	"stocks/notifications"
	"stocks/securities"
	"stocks/securities/direxion"
)

func orchestrate(ctx context.Context, db database.DB, client securities.Client, parsers []alerts.AlertParser, notifier notifications.Notifier) error {
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

	var gatheredAlerts []notifications.NotifierRequest
	for _, parser := range parsers {
		tAlerts, subscribers, err := parser.GetAlerts(ctx, holdingsMap)
		if err != nil {
			return err
		}
		gatheredAlerts = append(gatheredAlerts, notifications.NotifierRequest{
			Alerts:         tAlerts,
			Subscribers:    subscribers,
			Title:          "Notifications!!!",
			AlertGroupName: "Leveraged Stock Alerts",
		})
	}
	fmt.Printf("Found alerts: %+v\n", gatheredAlerts)

	if notifier != nil {
		for _, alert := range gatheredAlerts {
			_, err := notifier.Send(ctx, alert)
			if err != nil {
				return err
			}
		}
	}

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
		log.Fatal(err)
		return
	}

	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found Morning Star Config: %+v\n", config)

	msapi := morning_star.New(config.MSAPI)

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	var notifier notifications.Notifier
	if config.Notifications.ShouldSendEmails {
		notifier = notifications.New(notifications.Config{TempDirectory: "tmp"})
	}

	err = orchestrate(ctx, db, client, alertParsers, notifier)
	if err != nil {
		log.Fatal(err)
	}
}
