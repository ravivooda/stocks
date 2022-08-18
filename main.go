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
	"time"
)

func main() {
	defer utils.Elapsed("main")
	ctx, db, direxionClient, err, microSectorClient, etfs, config, insightsConfig, proSharesClient, alertParsers, notifier, masterdatareportsClient, logger, websitePaths := defaults()

	etfsMap := utils.MappedLETFS(etfs)
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
		etfsMaps:     etfsMap,
	})
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(totalHoldings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(totalHoldings)
	utils.PanicErr(err)

	shouldOrchestrate := true
	if shouldOrchestrate {
		orchestrate(ctx, orchestrateRequest{
			config:            config,
			parsers:           alertParsers,
			notifier:          notifier,
			insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)},
			insightsLogger:    logger,
			etfsMaps:          etfsMap,
		}, holdingsWithStockTickerMap, holdingsWithAccountTickerMap)
	}

	stocksMap := map[models.StockTicker]models.StockMetadata{}
	for stockTicker, holdings := range holdingsWithStockTickerMap {
		stocksMap[stockTicker] = models.StockMetadata{
			StockTicker:      stockTicker,
			StockDescription: holdings[0].StockDescription,
		}
	}

	providersMap := map[models.Provider]models.ProviderMetadata{}
	for ticker, holdings := range holdingsWithAccountTickerMap {
		provider := models.Provider(holdings[0].Provider)
		prev := providersMap[provider]
		prev.ETFTickers = append(prev.ETFTickers, ticker)
		providersMap[provider] = prev
	}

	metadata := website.Metadata{
		AccountMap:   holdingsWithAccountTickerMap,
		EtfsMap:      etfsMap,
		StocksMap:    stocksMap,
		ProvidersMap: providersMap,
	}
	testDuration := time.Duration(int64(config.Secrets.TestConfig.MaxServerRunTime)) * time.Second
	beginServing(ctx, insightsConfig, logger, websitePaths, metadata, testDuration)
}

func beginServing(
	ctx context.Context,
	insightsConfig insights.Config,
	logger insights.Logger,
	paths website.Paths,
	metadata website.Metadata,
	testDuration time.Duration,
) {
	server := website.New(website.Config{
		InsightsConfig: insightsConfig,
		WebsitePaths:   paths,
	}, logger, metadata)
	fmt.Println("started serving!!")
	utils.PanicErr(server.StartServing(ctx, testDuration))
	fmt.Println("done serving")
}

func defaults() (context.Context, database.DB, securities.Client, error, securities.Client, []models.ETF, Config, insights.Config, securities.SeedProvider, []alerts.AlertParser, notifications.Notifier, masterdatareports.Client, insights.Logger, website.Paths) {
	ctx := context.Background()
	db := database.NewDumbDatabase()
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
	direxionClient, err := direxion.NewClient(direxion.Config{TemporaryDir: config.Directories.Temporary})
	fmt.Printf("Found Morning Star Config: %+v\n", config)
	insightsConfig := insights.Config{
		OverlapsDirectory:    config.Directories.Artifacts + "/overlaps",
		ETFHoldingsDirectory: config.Directories.Artifacts + "/etf_holdings",
		StocksDirectory:      config.Directories.Artifacts + "/stocks",
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

	logger := insights.NewInsightsLogger(insightsConfig)

	websitePaths := website.DefaultWebsitePaths
	return ctx, db, direxionClient, err, microSectorClient, etfs, config, insightsConfig, proSharesClient, alertParsers, notifier, masterdatareportsClient, logger, websitePaths
}
