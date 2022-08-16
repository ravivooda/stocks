package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"stocks/alerts"
	"stocks/database"
	"stocks/database/insights"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/securities"
	"stocks/securities/masterdatareports"
	"stocks/utils"
	"stocks/website/letf"
)

type orchestrateRequest struct {
	config            Config
	parsers           []alerts.AlertParser
	notifier          notifications.Notifier
	insightGenerators []overlap.Generator
	insightsLogger    insights.Logger
	websiteGenerators []letf.Generator
	etfsMaps          map[models.LETFAccountTicker]models.ETF
}

type clientHoldingsRequest struct {
	config         Config
	etfs           []models.ETF
	seedGenerators []database.DB
	clients        map[models.Provider]securities.Client
	backupClient   masterdatareports.Client
	etfsMaps       map[models.LETFAccountTicker]models.ETF
}

func getHoldings(ctx context.Context, holdingsRequest clientHoldingsRequest) ([]models.LETFHolding, error) {
	var seeds []models.Seed
	for _, generator := range holdingsRequest.seedGenerators {
		_seeds, err := generator.ListSeeds(ctx)
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, _seeds...)
	}
	fmt.Printf("found %d seeds", len(seeds))

	clientHoldings, err := fetchHoldings(ctx, seeds, holdingsRequest.clients, holdingsRequest.etfsMaps)
	if err != nil {
		return nil, err
	}

	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(clientHoldings)

	var noMatchETFs []string
	fmt.Println("Loading master data reports client")
	for _, etf := range holdingsRequest.etfs {
		// first, try the normal client
		if _, ok := holdingsWithAccountTickerMap[etf.Symbol]; ok {
			continue
		}
		holdings, err := holdingsRequest.backupClient.GetHoldings(ctx, etf)
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].PercentContained > holdings[j].PercentContained
		})
		if err != nil {
			noMatchETFs = append(noMatchETFs, string(etf.Symbol))
			continue
		}
		holdingsWithAccountTickerMap[etf.Symbol] = holdings
	}
	fmt.Printf("did not find matching holdings for %+v (len: %d)\n", noMatchETFs, len(noMatchETFs))

	var totalHoldings []models.LETFHolding
	var (
		minPercentageTotal = 100.0
		maxPercentageTotal = 0.0
	)
	var mapsDidNotSumUp = map[string]float64{}
	for seed, holdings := range holdingsWithAccountTickerMap {
		sum := utils.SumHoldings(holdings)
		if math.Abs(sum-100) > 30 {
			//filteredHoldings := utils.FilteredForPrinting(holdings)
			//return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for etf %+v with holdings %+v", sum, seed, filteredHoldings))
			mapsDidNotSumUp[string(seed)] = sum
		}
		minPercentageTotal = math.Min(sum, minPercentageTotal)
		maxPercentageTotal = math.Max(sum, maxPercentageTotal)
		totalHoldings = append(totalHoldings, holdings...)
	}
	fmt.Printf("did not add up for: %v\n", mapsDidNotSumUp)
	fmt.Printf("Found minPercentageTotal: %f, maxPercentageTotal: %f\n", minPercentageTotal, maxPercentageTotal)

	return totalHoldings, nil
}

func orchestrate(
	ctx context.Context,
	request orchestrateRequest,
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
) {
	gatheredAlerts, err := gatherAlerts(ctx, request.parsers, holdingsWithStockTickerMap)
	utils.PanicErr(err)
	fmt.Printf("Found alerts: %d\n", len(gatheredAlerts))

	if request.config.Secrets.Notifications.ShouldSendEmails {
		_, err := request.notifier.SendAll(ctx, gatheredAlerts)
		utils.PanicErr(err)
	}

	logHoldings(ctx, request.insightsLogger, holdingsWithAccountTickerMap)

	if request.config.Secrets.Uploads.ShouldUploadInsightsOutputToGCP {
		generateInsights(ctx, request, holdingsWithAccountTickerMap, holdingsWithStockTickerMap)
	}
}

func logHoldings(ctx context.Context, logger insights.Logger, holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding) {
	for ticker, holdings := range holdingsWithAccountTickerMap {
		fileName, err := logger.LogHoldings(ctx, ticker, holdings)
		utils.PanicErr(err)
		fmt.Printf("wrote the holdings to %s\n", fileName)
	}
}

func generateInsights(
	ctx context.Context,
	request orchestrateRequest,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding,
) {
	var totalGatheredInsights = 0
	//fmt.Printf("Generating %d stock summaries\n", len(holdingsWithStockTickerMap))
	//for _, generator := range request.websiteGenerators {
	//	for stockTicker, holdings := range holdingsWithStockTickerMap {
	//		generator.GenerateStock(ctx, stockTicker, holdings)
	//	}
	//}
	//
	//fmt.Println("Generating welcome pages")
	//for _, generator := range request.websiteGenerators {
	//	_, err := generator.Generate(ctx, letf.Request{
	//		Letfs:     holdingsWithAccountTickerMap,
	//		StocksMap: holdingsWithStockTickerMap,
	//	})
	//	utils.PanicErr(err)
	//}

	fmt.Println("Generating ETF Pages")
	for _, iGenerator := range request.insightGenerators {
		iGenerator.Generate(holdingsWithAccountTickerMap, func(value []models.LETFOverlapAnalysis) {
			letfMappedOverlappedAnalysis := utils.MapLETFAnalysisWithAccountTicker(value)
			for _, letfOverlapAnalyses := range letfMappedOverlappedAnalysis {
				totalGatheredInsights += len(letfOverlapAnalyses)
				mappedOverlapAnalysis := map[string][]models.LETFOverlapAnalysis{}
				for _, analysis := range letfOverlapAnalyses {
					holdee := request.etfsMaps[analysis.LETFHoldees[0]]
					etfArray := mappedOverlapAnalysis[holdee.Leveraged]
					if etfArray == nil {
						etfArray = []models.LETFOverlapAnalysis{}
					}
					mappedOverlapAnalysis[holdee.Leveraged] = append(etfArray, analysis)
				}

				for leverage, overlapAnalyses := range mappedOverlapAnalysis {
					for _, analysis := range overlapAnalyses {
						_, err := request.insightsLogger.LogOverlapAnalysis(leverage, analysis)
						utils.PanicErr(err)
					}
				}
			}
		})
	}
	fmt.Printf("Total insights count: %d\n", totalGatheredInsights)
}

func gatherAlerts(
	ctx context.Context,
	parsers []alerts.AlertParser,
	holdingsMap map[models.StockTicker][]models.LETFHolding,
) ([]notifications.NotifierRequest, error) {
	var gatheredAlerts []notifications.NotifierRequest
	for _, parser := range parsers {
		tAlerts, subscribers, err := parser.GetAlerts(ctx, holdingsMap)
		if err != nil {
			return nil, err
		}
		gatheredAlerts = append(gatheredAlerts, notifications.NotifierRequest{
			Alerts:         tAlerts,
			Subscribers:    subscribers,
			Title:          "Notifications!!!",
			AlertGroupName: "Leveraged Stock Alerts",
		})
	}
	return gatheredAlerts, nil
}

func fetchHoldings(ctx context.Context, seeds []models.Seed, clientsMap map[models.Provider]securities.Client, maps map[models.LETFAccountTicker]models.ETF) ([]models.LETFHolding, error) {
	var allHoldings []models.LETFHolding
	for _, seed := range seeds {
		client := clientsMap[seed.Provider]
		if client == nil {
			return nil, errors.New(fmt.Sprintf("did not find provider for seed: %+v", seed))
		}
		fmt.Printf("fetching information for %+v\n", seed)
		holdings, err := client.GetHoldings(ctx, seed, maps[utils.FetchAccountTicker(seed.Ticker)])
		if err != nil {
			return nil, err
		}
		allHoldings = append(allHoldings, holdings...)
	}

	sort.Slice(allHoldings, func(i, j int) bool {
		return allHoldings[i].PercentContained > allHoldings[j].PercentContained
	})

	return allHoldings, nil
}
