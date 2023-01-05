package orchestrate

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"stocks/alerts"
	"stocks/database"
	"stocks/database/insights"
	"stocks/external/securities"
	"stocks/external/securities/masterdatareports"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/utils"
)

type Request struct {
	Config            Config
	Parsers           []alerts.AlertParser
	Notifier          notifications.Notifier
	InsightGenerators []overlap.Generator
	InsightsLogger    insights.Logger
	EtfsMaps          map[models.LETFAccountTicker]models.ETF
}

type ClientHoldingsRequest struct {
	Config         Config
	ETFs           []models.ETF
	SeedGenerators []database.DB
	Clients        map[models.Provider]securities.Client
	BackupClient   masterdatareports.Client
	EtfsMaps       map[models.LETFAccountTicker]models.ETF
}

func GetHoldings(ctx context.Context, holdingsRequest ClientHoldingsRequest) ([]models.LETFHolding, error) {
	var seeds []models.Seed
	for _, generator := range holdingsRequest.SeedGenerators {
		_seeds, err := generator.ListSeeds(ctx)
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, _seeds...)
	}
	fmt.Printf("found %d seeds", len(seeds))

	clientHoldings, err := fetchHoldings(ctx, seeds, holdingsRequest.Clients, holdingsRequest.EtfsMaps)
	if err != nil {
		return nil, err
	}

	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(clientHoldings)

	var noMatchETFs []string
	fmt.Println("Loading master data reports client")
	for _, etf := range holdingsRequest.ETFs {
		// first, try the normal client
		if _, ok := holdingsWithAccountTickerMap[etf.Symbol]; ok {
			continue
		}
		holdings, err := holdingsRequest.BackupClient.GetHoldings(ctx, etf)
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
	var mapsDidNotSumUp = map[string][]string{}
	for seed, holdings := range holdingsWithAccountTickerMap {
		sum := utils.SumHoldings(holdings)
		if math.Abs(sum-100) > 30 {
			//filteredHoldings := utils.FilteredForPrinting(holdings)
			//return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for etf %+v with holdings %+v", sum, seed, filteredHoldings))
			mapsDidNotSumUp[holdings[0].Provider] = append(mapsDidNotSumUp[holdings[0].Provider], fmt.Sprintf("%s, %f, provider: %s", seed, sum, holdings[0].Provider))
			delete(holdingsWithAccountTickerMap, seed)
		}
		minPercentageTotal = math.Min(sum, minPercentageTotal)
		maxPercentageTotal = math.Max(sum, maxPercentageTotal)
		totalHoldings = append(totalHoldings, holdings...)
	}
	fmt.Printf("did not add up for: %v\n", mapsDidNotSumUp)
	fmt.Printf("Found minPercentageTotal: %f, maxPercentageTotal: %f\n", minPercentageTotal, maxPercentageTotal)

	return totalHoldings, nil
}

func Orchestrate(ctx context.Context, request Request, holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding, holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding) int {
	gatheredAlerts, err := gatherAlerts(ctx, request.Parsers, holdingsWithStockTickerMap)
	utils.PanicErr(err)
	fmt.Printf("Found alerts: %d\n", len(gatheredAlerts))

	if request.Config.Secrets.Notifications.ShouldSendEmails {
		_, err := request.Notifier.SendAll(ctx, gatheredAlerts)
		utils.PanicErr(err)
	}

	logHoldings(ctx, request.InsightsLogger, holdingsWithAccountTickerMap, request.EtfsMaps)
	logStocks(ctx, request, holdingsWithStockTickerMap, holdingsWithAccountTickerMap, request.EtfsMaps)

	return generateInsights(ctx, request, holdingsWithAccountTickerMap)
}

func logHoldings(
	ctx context.Context,
	logger insights.Logger,
	holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding,
	etfsMap map[models.LETFAccountTicker]models.ETF,
) {
	for ticker, holdings := range holdingsWithAccountTickerMap {
		fileName, err := logger.LogHoldings(ctx, ticker, holdings, etfsMap[ticker].Leveraged)
		utils.PanicErr(err)
		fmt.Printf("wrote the holdings to %s\n", fileName)
	}
}

func generateInsights(_ context.Context, request Request, holdingsWithAccountTickerMap map[models.LETFAccountTicker][]models.LETFHolding) int {
	var totalGatheredInsights = 0

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
	for _, iGenerator := range request.InsightGenerators {
		iGenerator.Generate(holdingsWithAccountTickerMap, func(value []models.LETFOverlapAnalysis) {
			letfMappedOverlappedAnalysis := utils.MapLETFAnalysisWithAccountTicker(value)
			for ticker, letfOverlapAnalyses := range letfMappedOverlappedAnalysis {
				totalGatheredInsights += len(letfOverlapAnalyses)
				mappedOverlapAnalysis := map[string][]models.LETFOverlapAnalysis{}
				for _, analysis := range letfOverlapAnalyses {
					holdee := request.EtfsMaps[analysis.LETFHoldees[0]]
					etfArray := mappedOverlapAnalysis[holdee.Leveraged]
					if etfArray == nil {
						etfArray = []models.LETFOverlapAnalysis{}
					}
					mappedOverlapAnalysis[holdee.Leveraged] = append(etfArray, analysis)
				}

				var wrappedOverlaps []insights.OverlapWrapper
				for leverage, overlapAnalyses := range mappedOverlapAnalysis {
					for _, analysis := range overlapAnalyses {
						wrappedOverlaps = append(wrappedOverlaps, insights.OverlapWrapper{
							Leverage: leverage,
							Analysis: models.LETFOverlapAnalysis{
								LETFHolder:        analysis.LETFHolder,
								LETFHoldees:       analysis.LETFHoldees,
								OverlapPercentage: analysis.OverlapPercentage,
								DetailedOverlap:   nil,
							},
						})
					}
				}

				_, err := request.InsightsLogger.LogOverlapAnalysisForHolder(ticker, wrappedOverlaps)
				utils.PanicErr(err)
			}
		})
	}
	fmt.Printf("Total insights count: %d\n", totalGatheredInsights)
	return totalGatheredInsights
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
