package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
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

type orchestrateRequest struct {
	config            Config
	seedGenerators    []database.DB
	clients           map[models.Provider]securities.Client
	parsers           []alerts.AlertParser
	notifier          notifications.Notifier
	insightGenerators []overlap.Generator
	insightsLogger    insights.Logger
	websiteGenerators []letf.Generator
}

func orchestrateV1(ctx context.Context, request orchestrateRequest) error {
	var seeds []models.Seed
	for _, generator := range request.seedGenerators {
		_seeds, err := generator.ListSeeds(ctx)
		if err != nil {
			return err
		}
		seeds = append(seeds, _seeds...)
	}
	fmt.Printf("found %d seeds", len(seeds))

	holdings, err := fetchHoldings(ctx, seeds, request.clients)
	if err != nil {
		return err
	}

	return orchestrate(ctx, request, holdings)
}

func orchestrate(ctx context.Context, request orchestrateRequest, holdings []models.LETFHolding) error {
	holdingsWithStockTickerMap := utils.MapLETFHoldingsWithStockTicker(holdings)
	holdingsWithAccountTickerMap := utils.MapLETFHoldingsWithAccountTicker(holdings)
	//fmt.Println(holdingsWithStockTickerMap)

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

	analysisMap, totalInsightsCount, err := gatherInsights(ctx, request.insightGenerators, holdingsWithAccountTickerMap)
	if err != nil {
		return err
	}
	fmt.Printf("Total insights count: %d\n", totalInsightsCount)

	for _, analysis := range analysisMap {
		for _, insight := range analysis {
			_, err := request.insightsLogger.Log(insight)
			if err != nil {
				return err
			}
		}
	}

	for _, generator := range request.websiteGenerators {
		_, err := generator.Generate(ctx, letf.Request{
			AnalysisMap: analysisMap,
			Letfs:       holdingsWithAccountTickerMap,
			StocksMap:   holdingsWithStockTickerMap,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func gatherInsights(_ context.Context, generators []overlap.Generator, letfHoldings map[models.LETFAccountTicker][]models.LETFHolding) (map[models.LETFAccountTicker][]models.LETFOverlapAnalysis, int, error) {
	var mapGatheredInsights = map[models.LETFAccountTicker][]models.LETFOverlapAnalysis{}
	var totalGatheredInsights = 0
	for _, generator := range generators {
		overlapAnalyses := generator.Generate(letfHoldings)
		//mergedAnalyses := generator.MergeInsights(overlapAnalyses, letfHoldings)
		//for ticker, analyses := range mergedAnalyses {
		//	overlapAnalyses[ticker] = append(overlapAnalyses[ticker], analyses...)
		//}
		for ticker, analyses := range overlapAnalyses {
			mapGatheredInsights[ticker] = append(mapGatheredInsights[ticker], analyses...)
			totalGatheredInsights += len(analyses)
		}
	}
	return mapGatheredInsights, totalGatheredInsights, nil
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

func fetchHoldings(
	ctx context.Context,
	seeds []models.Seed,
	clientsMap map[models.Provider]securities.Client,
) ([]models.LETFHolding, error) {
	var allHoldings []models.LETFHolding
	for _, seed := range seeds {
		client := clientsMap[seed.Provider]
		if client == nil {
			return nil, errors.New(fmt.Sprintf("did not find provider for seed: %+v", seed))
		}
		fmt.Printf("fetching information for %+v\n", seed)
		holdings, err := client.GetHoldings(ctx, seed)
		if err != nil {
			return nil, err
		}

		if sum := utils.SumHoldings(holdings); math.Abs(sum-100) > 0.5 {
			return nil, errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for seed %+v", sum, seed))
		}
		allHoldings = append(allHoldings, holdings...)
	}

	sort.Slice(allHoldings, func(i, j int) bool {
		return allHoldings[i].PercentContained > allHoldings[j].PercentContained
	})

	return allHoldings, nil
}

func oldMain() {
	ctx := context.Background()
	db := database.NewDumbDatabase()
	direxionClient, err := direxion.NewClient()
	microSectorClient, err := microsector.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
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
	err = orchestrateV1(ctx, orchestrateRequest{
		config: config,
		seedGenerators: []database.DB{
			db,
			proSharesClient,
		},
		clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microSectorClient,
			models.ProShares:   proSharesClient,
		},
		parsers:           alertParsers,
		notifier:          notifier,
		insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)},
		insightsLogger:    insights.NewInsightsLogger(insights.Config{RootDir: config.Directories.Artifacts + "/insights"}),
		websiteGenerators: []letf.Generator{letf.New(letf.Config{WebsiteDirectoryRoot: config.Directories.Websites, MinThreshold: config.Outputs.Websites.MinThresholdPercentage})},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx := context.Background()
	etfsGenerator := etfdb.New(etfdb.Config{})
	etfs, err := etfsGenerator.ListETFs(ctx)
	if err != nil {
		return
	}
	fmt.Printf("Found %d etfs\n", len(etfs))

	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found Morning Star Config: %+v\n", config)

	msapi := morning_star.New(config.MSAPI)

	alertParsers := []alerts.AlertParser{
		movers.New(movers.Config{MSAPI: msapi}),
	}

	notifier := notifications.New(notifications.Config{TempDirectory: config.Directories.Temporary})

	fmt.Println("Loading master data reports client")
	var totalHoldings []models.LETFHolding
	masterdatareportsClient, err := masterdatareports.New(config.Securities.MasterDataReports)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded master data reports client, found %d number of etfs data\n", masterdatareportsClient.Count())
	var noMatchETFs []string
	for _, etf := range etfs {
		holdings, err := masterdatareportsClient.GetHoldings(ctx, etf)
		if err != nil {
			noMatchETFs = append(noMatchETFs, string(etf.Symbol))
			continue
		}
		if sum := utils.SumHoldings(holdings); math.Abs(sum-100) > 30 {
			filteredHoldings := utils.FilteredForPrinting(holdings)
			panic(errors.New(fmt.Sprintf("total percentage (%f) did not add up to 100 percent for etf %+v with holdings %+v", sum, etf, filteredHoldings)))
		}
		totalHoldings = append(totalHoldings, holdings...)
	}
	fmt.Printf("did not find matching holdings for %+v (len: %d)\n", noMatchETFs, len(noMatchETFs))

	err = orchestrate(ctx, orchestrateRequest{
		config:            config,
		seedGenerators:    nil,
		clients:           nil,
		parsers:           alertParsers,
		notifier:          notifier,
		insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(config.Outputs.Insights)},
		insightsLogger:    insights.NewInsightsLogger(insights.Config{RootDir: config.Directories.Artifacts + "/insights"}),
		websiteGenerators: []letf.Generator{letf.New(letf.Config{WebsiteDirectoryRoot: config.Directories.Websites, MinThreshold: config.Outputs.Websites.MinThresholdPercentage})},
	}, totalHoldings)
	if err != nil {
		panic(err)
	}
}
