package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"sort"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/database/insights"
	"stocks/external/securities"
	"stocks/external/securities/direxion"
	"stocks/external/securities/invesco"
	"stocks/external/securities/masterdatareports"
	"stocks/external/securities/microsector"
	"stocks/external/securities/proshares"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/orchestrate"
	"stocks/utils"
	"stocks/website"
)

type setupRequest struct {
	metadataFileDestination     string
	autoCompleteFileDestination string
	config                      orchestrate.Config
	insightsConfig              insights.Config
	db                          database.DB
	etfs                        []models.ETF
	generators                  []overlap.Generator
	logger                      insights.Logger
}

func setup(context context.Context, shouldOrchestrate bool, request setupRequest) {
	defer utils.Elapsed("setup")()
	etfsMap := utils.MappedLETFS(request.etfs)
	microSectorClient, direxionClient, proSharesClient, masterdatareportsClient, alertParsers, notifier, invescoClient := createSecurityClients(request.config, etfsMap)

	clientHoldingsRequest := orchestrate.ClientHoldingsRequest{
		Config: request.config,
		ETFs:   request.etfs,
		SeedGenerators: []database.DB{
			request.db,
			proSharesClient,
			invescoClient,
		},
		Clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microSectorClient,
			models.ProShares:   proSharesClient,
			models.Invesco:     invescoClient,
		},
		BackupClient: masterdatareportsClient,
		EtfsMaps:     etfsMap,
	}
	totalHoldings, err := orchestrate.GetHoldings(context, clientHoldingsRequest)
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(totalHoldings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(totalHoldings)
	utils.PanicErr(err)
	generatedInsights := 600000
	if shouldOrchestrate {
		generatedInsights = orchestrate.Orchestrate(context, orchestrate.Request{
			Config:            request.config,
			Parsers:           alertParsers,
			Notifier:          notifier,
			InsightGenerators: request.generators,
			InsightsLogger:    request.logger,
			EtfsMaps:          etfsMap,
		}, holdingsWithStockTickerMap, holdingsWithAccountTickerMap, func(ticker models.LETFAccountTicker) bool {
			return true
		})
	}

	metadata, autoCompleteMetadata := createMetadata(holdingsWithStockTickerMap, holdingsWithAccountTickerMap, etfsMap, 10, 10, generatedInsights)
	// Write metadata
	b, err := json.Marshal(metadata)
	utils.PanicErr(err)
	utils.PanicErr(ioutil.WriteFile(request.metadataFileDestination, b, fs.ModePerm))
	// Write autocomplete metadata
	c, err := json.Marshal(autoCompleteMetadata)
	utils.PanicErr(err)
	utils.PanicErr(ioutil.WriteFile(request.autoCompleteFileDestination, c, fs.ModePerm))
}

func createMetadata(
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfsMap map[models.LETFAccountTicker]models.ETF,
	topStocksCount int,
	topETFsCount int,
	generatedInsights int,
) (website.Metadata, website.AutoCompleteMetadata) {
	stocksMap, providersMap, accountMap := createMaps(holdingsWithStockTickerMap, holdingsWithAccountTickerMap, etfsMap)

	topStocks := filterTopStocks(stocksMap, topStocksCount)

	// TODO: Bring back hardcoded top ETFS; Today it's hardcoded as top etfs by volume trading from: https://etfdb.com/compare/volume/
	//topETFs := filterTopETFs(accountMap, topETFsCount)
	topETFs := TopHardcodedETFs

	metadata := website.Metadata{
		AccountMap:   accountMap,
		StocksMap:    stocksMap,
		ProvidersMap: providersMap,
		TemplateCustomMetadata: website.TemplateCustomMetadata{
			SideBarMetadata: website.SideBarMetadata{
				TopETFs:   topETFs,
				TopStocks: topStocks,
				SocialNetworkMetadata: website.SocialNetworkMetadata{ //TODO: Should move this to configuration instead.
					LinkedInURL: "https://tinyurl.com/tlhtli",
					FacebookURL: "https://tinyurl.com/tlhtfb",
					TwitterURL:  "https://tinyurl.com/tlhttw",
				},
			},
		},
		GeneratedInsightsCount: generatedInsights,
	}
	return metadata, website.AutoCompleteMetadata{
		StocksMap:  utils.MapToArrayForStockTickers(stocksMap),
		AccountMap: utils.MapToArrayForAccountTickers(accountMap),
	}
}

func filterTopETFs(accountMap map[models.LETFAccountTicker]models.ETFMetadata, topETFsCount int) []models.LETFAccountTicker {
	var topETFsMetadata []models.ETFMetadata
	for _, metadata := range accountMap {
		topETFsMetadata = append(topETFsMetadata, metadata)
	}
	sort.Slice(topETFsMetadata, func(i, j int) bool {
		return topETFsMetadata[i].HoldingsCount > topETFsMetadata[j].HoldingsCount
	})
	topETFsMetadata = topETFsMetadata[:topETFsCount]
	fmt.Printf("Top ETFs Metadata: %+v\n", topETFsMetadata)

	var topETFs []models.LETFAccountTicker
	for _, metadatum := range topETFsMetadata {
		topETFs = append(topETFs, metadatum.Ticker)
	}
	return topETFs
}

func filterTopStocks(stocksMap map[models.StockTicker]models.StockMetadata, topStocksCount int) []models.StockTicker {
	var topStocksMetadata []models.StockMetadata
	for _, metadata := range stocksMap {
		topStocksMetadata = append(topStocksMetadata, metadata)
	}
	sort.Slice(topStocksMetadata, func(i, j int) bool {
		return topStocksMetadata[i].ETFCount > topStocksMetadata[j].ETFCount
	})
	topStocksMetadata = topStocksMetadata[:topStocksCount]
	fmt.Printf("Top Stocks Metadata: %+v\n", topStocksMetadata)

	var topStocks []models.StockTicker
	for _, stock := range topStocksMetadata {
		topStocks = append(topStocks, stock.StockTicker)
	}
	return topStocks
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
		m := map[models.LETFAccountTicker]bool{}
		for _, holding := range holdings {
			m[holding.LETFAccountTicker] = true
		}
		stocksMap[stockTicker] = models.StockMetadata{
			StockTicker:      stockTicker,
			StockDescription: holdings[0].StockDescription,
			ETFCount:         len(m),
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
		leveraged := etfsMap[ticker].Leveraged
		if leveraged == "" {
			panic(fmt.Sprintf("found empty leverage for etf: %s", ticker))
		}
		accountMap[ticker] = models.ETFMetadata{
			Ticker:        ticker,
			Provider:      models.Provider(holdings[0].Provider),
			Description:   holdings[0].LETFDescription,
			Leveraged:     leveraged,
			HoldingsCount: len(holdings),
		}
	}
	return stocksMap, providersMap, accountMap
}

func createSecurityClients(config orchestrate.Config, etfsMap map[models.LETFAccountTicker]models.ETF) (
	microsectorClient securities.Client,
	direxionClient securities.Client,
	prosharesClient securities.SeedProvider,
	mdrClient masterdatareports.Client,
	parser []alerts.AlertParser,
	notifier notifications.Notifier,
	invescoClient securities.SeedProvider,
) {
	microSectorClient, err := microsector.NewClient()
	utils.PanicErr(err)
	direxionClient, err = direxion.NewClient(direxion.Config{TemporaryDir: config.Directories.Temporary})
	utils.PanicErr(err)
	proSharesClient, err := proshares.New(config.Securities.ProShares)
	utils.PanicErr(err)
	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports, etfsMap)
	utils.PanicErr(err)
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())

	msapi := morning_star.New(config.MSAPI)
	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier = notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	invescoClient = invesco.New(config.Securities.Invesco)

	return microSectorClient, direxionClient, proSharesClient, masterdatareportsClient, alertParsers, notifier, invescoClient
}
