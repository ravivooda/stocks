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
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/orchestrate"
	"stocks/securities"
	"stocks/securities/direxion"
	"stocks/securities/masterdatareports"
	"stocks/securities/microsector"
	"stocks/securities/proshares"
	"stocks/utils"
	"stocks/website"
)

type setupRequest struct {
	fileAddr       string
	config         orchestrate.Config
	insightsConfig insights.Config
	db             database.DB
	etfs           []models.ETF
	generators     []overlap.Generator
	logger         insights.Logger
}

func setup(context context.Context, shouldOrchestrate bool, request setupRequest) {
	defer utils.Elapsed("setup")()
	microSectorClient, direxionClient, proSharesClient, masterdatareportsClient, alertParsers, notifier := createSecurityClients(request.config)

	etfsMap := utils.MappedLETFS(request.etfs)
	clientHoldingsRequest := orchestrate.ClientHoldingsRequest{
		Config: request.config,
		ETFs:   request.etfs,
		SeedGenerators: []database.DB{
			request.db,
			proSharesClient,
		},
		Clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microSectorClient,
			models.ProShares:   proSharesClient,
		},
		BackupClient: masterdatareportsClient,
		EtfsMaps:     etfsMap,
	}
	totalHoldings, err := orchestrate.GetHoldings(context, clientHoldingsRequest)
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(totalHoldings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(totalHoldings)
	utils.PanicErr(err)

	if shouldOrchestrate {
		orchestrate.Orchestrate(context, orchestrate.Request{
			Config:            request.config,
			Parsers:           alertParsers,
			Notifier:          notifier,
			InsightGenerators: request.generators,
			InsightsLogger:    request.logger,
			EtfsMaps:          etfsMap,
		}, holdingsWithStockTickerMap, holdingsWithAccountTickerMap)
	}

	metadata := createMetadata(holdingsWithStockTickerMap, holdingsWithAccountTickerMap, etfsMap, 10, 10)
	b, err := json.Marshal(metadata)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(request.fileAddr, b, fs.ModePerm))
}

func createMetadata(
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfsMap map[models.LETFAccountTicker]models.ETF,
	topStocksCount int,
	topETFsCount int,
) website.Metadata {
	stocksMap, providersMap, accountMap := createMaps(holdingsWithStockTickerMap, holdingsWithAccountTickerMap, etfsMap)

	topStocks := filterTopStocks(stocksMap, topStocksCount)

	topETFs := filterTopETFs(accountMap, topETFsCount)

	metadata := website.Metadata{
		AccountMap:   accountMap,
		StocksMap:    stocksMap,
		ProvidersMap: providersMap,
		TemplateCustomMetadata: website.TemplateCustomMetadata{
			SideBarMetadata: website.SideBarMetadata{
				TopETFs:   topETFs,
				TopStocks: topStocks,
			},
		},
	}
	return metadata
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
		accountMap[ticker] = models.ETFMetadata{
			Ticker:        ticker,
			Provider:      models.Provider(holdings[0].Provider),
			Description:   holdings[0].LETFDescription,
			Leveraged:     etfsMap[ticker].Leveraged,
			HoldingsCount: len(holdings),
		}
	}
	return stocksMap, providersMap, accountMap
}

func createSecurityClients(config orchestrate.Config) (securities.Client, securities.Client, securities.SeedProvider, masterdatareports.Client, []alerts.AlertParser, notifications.Notifier) {
	microSectorClient, err := microsector.NewClient()
	utils.PanicErr(err)
	direxionClient, err := direxion.NewClient(direxion.Config{TemporaryDir: config.Directories.Temporary})
	utils.PanicErr(err)
	proSharesClient, err := proshares.New(config.Securities.ProShares)
	utils.PanicErr(err)
	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports)
	utils.PanicErr(err)
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())

	msapi := morning_star.New(config.MSAPI)
	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier := notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	return microSectorClient, direxionClient, proSharesClient, masterdatareportsClient, alertParsers, notifier
}
