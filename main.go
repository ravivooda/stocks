package main

import (
	"context"
	"fmt"
	"log"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/database/insights"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/securities"
	"stocks/securities/direxion"
	"stocks/securities/microsector"
	"stocks/securities/proshares"
	"stocks/website/letf"
)

func main() {
	ctx := context.Background()
	db := database.NewDumbDatabase()
	direxionClient, err := direxion.NewClient()
	microSectorClient, err := microsector.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	config, err := NewConfig()
	if err != nil {
		log.Fatalf("error occurred loading config: %+v \n", err)
	}
	fmt.Printf("Found Morning Star Config: %+v\n", config)

	msapi := morning_star.New(config.MSAPI)

	proSharesClient, err := proshares.New(config.Securities.ProShares)
	if err != nil {
		log.Fatal(err)
	}

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier := notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	totalHoldings, err := getHoldings(ctx, clientHoldingsRequest{
		config: config,
		etfs:   []models.ETF{},
		seedGenerators: []database.DB{
			db,
			proSharesClient,
		},
		clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microSectorClient,
			models.ProShares:   proSharesClient,
		},
	})
	if err != nil {
		panic(err)
	}

	generator, err := letf.New(letf.Config{WebsiteDirectoryRoot: config.Directories.Websites, MinThreshold: config.Outputs.Websites.MinThresholdPercentage})
	if err != nil {
		panic(err)
	}

	err = orchestrate(ctx, orchestrateRequest{
		config:            config,
		parsers:           alertParsers,
		notifier:          notifier,
		insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)},
		insightsLogger:    insights.NewInsightsLogger(insights.Config{RootDir: config.Directories.Artifacts + "/insights"}),
		websiteGenerators: []letf.Generator{generator},
	}, totalHoldings)
	if err != nil {
		panic(err)
	}
}
