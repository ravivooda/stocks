package main

import (
	"context"
	"fmt"
	"log"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/database/etfdb"
	"stocks/database/insights"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/securities"
	"stocks/securities/direxion"
	"stocks/securities/masterdatareports"
	"stocks/securities/microsector"
	"stocks/securities/proshares"
	"stocks/utils"
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

	etfsGenerator := etfdb.New(etfdb.Config{})
	etfs, err := etfsGenerator.ListETFs(ctx)
	if err != nil {
		return
	}
	fmt.Printf("Found %d etfs\n", len(etfs))

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

	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())

	totalHoldings, err := getHoldings(ctx, clientHoldingsRequest{
		config: config,
		etfs:   etfs,
		seedGenerators: []database.DB{
			db,
			proSharesClient,
		},
		clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microSectorClient,
			models.ProShares:   proSharesClient,
		},
		backupClient: masterdatareportsClient,
		etfsMaps:     utils.MappedLETFS(etfs),
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
		etfsMaps:          utils.MappedLETFS(etfs),
	}, totalHoldings)
	if err != nil {
		panic(err)
	}
}
