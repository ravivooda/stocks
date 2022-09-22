package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
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
	ctx, _, _, config, insightsConfig, _, _, websitePaths := defaults()
	fileAddr := fmt.Sprintf("%s/%s.json", insightsConfig.RootDirectory, "metadata")
	generators, logger := logicProviders(config, insightsConfig)
	if config.Secrets.Uploads.ShouldUploadInsightsOutputToGCP {
		setup(ctx, true, fileAddr, config)
	} else {
		// TODO: Hardcoded 0 index lookup on generators[0] in the line below
		serve(config, ctx, insightsConfig, generators[0], logger, websitePaths, fileAddr)
	}
}

func setup(context context.Context, shouldOrchestrate bool, fileAddr string, config Config) {
	defer utils.Elapsed("setup")()
	microSectorClient, direxionClient, proSharesClient, masterdatareportsClient := createSecurityClients(config)
	_, db, etfs, _, insightsConfig, alertParsers, notifier, _ := defaults()

	etfsMap := utils.MappedLETFS(etfs)
	totalHoldings, err := getHoldings(context, clientHoldingsRequest{
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

	generators, logger := logicProviders(config, insightsConfig)

	if shouldOrchestrate {
		orchestrate(context, orchestrateRequest{
			config:            config,
			parsers:           alertParsers,
			notifier:          notifier,
			insightGenerators: generators,
			insightsLogger:    logger,
			etfsMaps:          etfsMap,
		}, holdingsWithStockTickerMap, holdingsWithAccountTickerMap)
	}

	stocksMap, providersMap, accountMap := createMaps(holdingsWithStockTickerMap, holdingsWithAccountTickerMap, etfsMap)

	metadata := website.Metadata{
		AccountMap:   accountMap,
		StocksMap:    stocksMap,
		ProvidersMap: providersMap,
	}
	b, err := json.Marshal(metadata)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
}

func logicProviders(config Config, insightsConfig insights.Config) ([]overlap.Generator, insights.Logger) {
	return []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)}, insights.NewInsightsLogger(insightsConfig)
}

func serve(
	config Config,
	ctx context.Context,
	insightsConfig insights.Config,
	generator overlap.Generator,
	logger insights.Logger,
	websitePaths website.Paths,
	fileAddr string,
) {
	file, err := ioutil.ReadFile(fileAddr)
	utils.PanicErr(err)

	metadata := website.Metadata{}

	utils.PanicErr(json.Unmarshal(file, &metadata))
	testDuration := time.Duration(int64(config.Secrets.TestConfig.MaxServerRunTime)) * time.Second
	beginServing(ctx, insightsConfig, logger, generator, websitePaths, metadata, testDuration)
}

func createMaps(
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfsMap map[models.LETFAccountTicker]models.ETF,
) (
	map[models.StockTicker]models.StockMetadata,
	map[models.Provider]models.ProviderMetadata,
	map[models.LETFAccountTicker]models.ETFMetadata,
) {
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

	var accountMap = map[models.LETFAccountTicker]models.ETFMetadata{}
	for ticker, holdings := range holdingsWithAccountTickerMap {
		accountMap[ticker] = models.ETFMetadata{
			Provider:      models.Provider(holdings[0].Provider),
			Description:   holdings[0].LETFDescription,
			Leveraged:     etfsMap[ticker].Leveraged,
			HoldingsCount: len(holdings),
		}
	}
	return stocksMap, providersMap, accountMap
}

func beginServing(
	ctx context.Context,
	insightsConfig insights.Config,
	logger insights.Logger,
	generator overlap.Generator,
	paths website.Paths,
	metadata website.Metadata,
	testDuration time.Duration,
) {
	server := website.New(website.Config{
		InsightsConfig: insightsConfig,
		TemplateCustomMetadata: website.TemplateCustomMetadata{
			SideBarMetadata: website.SideBarMetadata{
				TopETFs: []models.LETFAccountTicker{
					"SPY",
					"TECL",
				},
				TopStocks: []models.StockTicker{
					"AAPL",
					"MSFT",
					"PYPL",
				},
			},
			WebsitePaths: paths,
		},
	}, website.Dependencies{
		Logger:    logger,
		Generator: generator,
	}, metadata)
	fmt.Println("started serving!!")
	utils.PanicErr(server.StartServing(ctx, testDuration))
	fmt.Println("done serving")
}

func defaults() (context.Context, database.DB, []models.ETF, Config, insights.Config, []alerts.AlertParser, notifications.Notifier, website.Paths) {
	ctx := context.Background()
	db := database.NewDumbDatabase()

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
		OverlapsDirectory:    config.Directories.Artifacts + "/overlaps",
		ETFHoldingsDirectory: config.Directories.Artifacts + "/etf_holdings",
		StocksDirectory:      config.Directories.Artifacts + "/stocks",
		RootDirectory:        config.Directories.Artifacts,
	}

	msapi := morning_star.New(config.MSAPI)

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier := notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	websitePaths := website.DefaultWebsitePaths
	return ctx, db, etfs, config, insightsConfig, alertParsers, notifier, websitePaths
}

func createSecurityClients(config Config) (securities.Client, securities.Client, securities.SeedProvider, masterdatareports.Client) {
	microSectorClient, err := microsector.NewClient()
	utils.PanicErr(err)
	direxionClient, err := direxion.NewClient(direxion.Config{TemporaryDir: config.Directories.Temporary})
	utils.PanicErr(err)
	proSharesClient, err := proshares.New(config.Securities.ProShares)
	utils.PanicErr(err)
	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports)
	utils.PanicErr(err)
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())
	return microSectorClient, direxionClient, proSharesClient, masterdatareportsClient
}
