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
	"stocks/website"
	"stocks/website/letf"
)

func main() {
	defer utils.Elapsed("main")
	ctx, db, direxionClient, err, microSectorClient, etfs, config, insightsConfig, proSharesClient, alertParsers, notifier, masterdatareportsClient, generator, logger, websitePaths := defaults()

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
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(totalHoldings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(totalHoldings)
	utils.PanicErr(err)

	shouldOrchestrate := false
	if shouldOrchestrate {
		orchestrate(ctx, orchestrateRequest{
			config:            config,
			parsers:           alertParsers,
			notifier:          notifier,
			insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)},
			insightsLogger:    logger,
			websiteGenerators: []letf.Generator{generator},
			etfsMaps:          utils.MappedLETFS(etfs),
		}, holdingsWithStockTickerMap, holdingsWithAccountTickerMap)
	}

	beginServing(ctx, insightsConfig, logger, websitePaths, holdingsWithAccountTickerMap)
}

func beginServing(
	ctx context.Context,
	insightsConfig insights.Config,
	logger insights.Logger,
	paths letf.WebsitePaths,
	tickerMap map[models.LETFAccountTicker][]models.LETFHolding,
) {
	server := website.New(website.Config{
		InsightsConfig: insightsConfig,
		WebsitePaths:   paths,
	}, logger, tickerMap)
	fmt.Println("started serving!!")
	utils.PanicErr(server.StartServing(ctx))
	fmt.Println("done serving")
}

func defaults() (context.Context, database.DB, securities.Client, error, securities.Client, []models.ETF, Config, insights.Config, securities.SeedProvider, []alerts.AlertParser, notifications.Notifier, masterdatareports.Client, letf.Generator, insights.Logger, letf.WebsitePaths) {
	ctx := context.Background()
	db := database.NewDumbDatabase()
	direxionClient, err := direxion.NewClient()
	microSectorClient, err := microsector.NewClient()
	utils.PanicErr(err)

	etfsGenerator := etfdb.New(etfdb.Config{})
	etfs, err := etfsGenerator.ListETFs(ctx)
	utils.PanicErr(err)
	fmt.Printf("Found %d etfs\n", len(etfs))

	config, err := NewConfig()
	if err != nil {
		log.Fatalf("error occurred loading config: %+v \n", err)
	}
	fmt.Printf("Found Morning Star Config: %+v\n", config)
	insightsConfig := insights.Config{
		OverlapsDirectory:    config.Directories.Artifacts + "/insights",
		ETFHoldingsDirectory: config.Directories.Artifacts + "/etf_holdings",
	}

	msapi := morning_star.New(config.MSAPI)

	proSharesClient, err := proshares.New(config.Securities.ProShares)
	utils.PanicErr(err)

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier := notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports)
	utils.PanicErr(err)
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())

	generator, err := letf.New(letf.Config{WebsiteDirectoryRoot: config.Directories.Websites, MinThreshold: config.Outputs.Websites.MinThresholdPercentage})
	utils.PanicErr(err)

	logger := insights.NewInsightsLogger(insightsConfig)

	websitePaths := letf.DefaultWebsitePaths
	return ctx, db, direxionClient, err, microSectorClient, etfs, config, insightsConfig, proSharesClient, alertParsers, notifier, masterdatareportsClient, generator, logger, websitePaths
}
