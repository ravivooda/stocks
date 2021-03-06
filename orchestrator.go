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

	var totalHoldings []models.LETFHolding
	for seed, holdings := range holdingsWithAccountTickerMap {
		if sum := utils.SumHoldings(holdings); math.Abs(sum-100) > 30 {
			filteredHoldings := utils.FilteredForPrinting(holdings)
			return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for etf %+v with holdings %+v", sum, seed, filteredHoldings))
		}
		totalHoldings = append(totalHoldings, holdings...)
	}

	return totalHoldings, nil
}

func orchestrate(ctx context.Context, request orchestrateRequest, holdings []models.LETFHolding) error {
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(holdings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(holdings)

	gatheredAlerts, err := gatherAlerts(ctx, request.parsers, holdingsWithStockTickerMap)
	if err != nil {
		return err
	}
	fmt.Printf("Found alerts: %d\n", len(gatheredAlerts))

	if request.config.Secrets.Notifications.ShouldSendEmails {
		_, err := request.notifier.SendAll(ctx, gatheredAlerts)
		if err != nil {
			return err
		}
	}

	if !request.config.Secrets.Uploads.ShouldUploadInsightsOutputToGCP {
		return nil
	}

	var totalGatheredInsights = 0
	for _, iGenerator := range request.insightGenerators {
		iGenerator.Generate(holdingsWithAccountTickerMap, func(value []models.LETFOverlapAnalysis) {
			letfMappedOverlappedAnalysis := utils.MapLETFAnalysisWithAccountTicker(value)
			for ticker, letfOverlapAnalyses := range letfMappedOverlappedAnalysis {
				totalGatheredInsights += len(letfOverlapAnalyses)
				for _, generator := range request.websiteGenerators {
					mappedOverlapAnalysis := map[string][]models.LETFOverlapAnalysis{}
					for _, analysis := range letfOverlapAnalyses {
						holdee := request.etfsMaps[analysis.LETFHoldees[0]]
						etfArray := mappedOverlapAnalysis[holdee.Leveraged]
						if etfArray == nil {
							etfArray = []models.LETFOverlapAnalysis{}
						}
						mappedOverlapAnalysis[holdee.Leveraged] = append(etfArray, analysis)
					}

					for leverage, analyses := range mappedOverlapAnalysis {
						mergedInsights := iGenerator.MergeInsights(map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{ticker: analyses}, holdingsWithAccountTickerMap)
						analyses = append(analyses, mergedInsights[ticker]...)
						mappedOverlapAnalysis[leverage] = analyses
					}
					_, err := generator.GenerateETF(ctx, ticker, mappedOverlapAnalysis, holdingsWithAccountTickerMap, holdingsWithStockTickerMap)
					if err != nil {
						panic(err)
					}
				}

				for _, insight := range letfOverlapAnalyses {
					_, err := request.insightsLogger.Log(insight)
					if err != nil {
						panic(err)
					}
				}
			}
		})
	}
	fmt.Printf("Total insights count: %d\n", totalGatheredInsights)
	return err
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
