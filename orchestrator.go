package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"stocks/alerts"
	"stocks/alerts/movers"
	"stocks/alerts/movers/morning_star"
	"stocks/database"
	"stocks/database/insights"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/notifications"
	"stocks/securities"
	"stocks/securities/direxion"
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

func orchestrate(ctx context.Context, request orchestrateRequest) error {
	var seeds []models.Seed
	for _, generator := range request.seedGenerators {
		_seeds, err := generator.ListSeeds(ctx)
		if err != nil {
			return err
		}
		seeds = append(seeds, _seeds...)
	}

	holdings, err := fetchHoldings(ctx, seeds, request.clients)
	if err != nil {
		return err
	}

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

	gatheredInsights, err := gatherInsights(ctx, request.insightGenerators, holdingsWithAccountTickerMap)
	if err != nil {
		return err
	}
	fmt.Printf("Total insights count: %d\n", len(gatheredInsights))

	for _, insight := range gatheredInsights {
		_, err := request.insightsLogger.Log(insight)
		if err != nil {
			return err
		}
		//fmt.Printf("Logged %s\n", fileName)
	}

	analysisMap := utils.MapLETFAnalysisWithAccountTicker(gatheredInsights)
	for _, generator := range request.websiteGenerators {
		_, err := generator.Generate(ctx, analysisMap, holdingsWithAccountTickerMap, holdingsWithStockTickerMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func gatherInsights(_ context.Context, generators []overlap.Generator, letfHoldings map[models.LETFAccountTicker][]models.LETFHolding) ([]models.LETFOverlapAnalysis, error) {
	var gatheredInsights []models.LETFOverlapAnalysis
	for _, generator := range generators {
		gatheredInsights = append(gatheredInsights, generator.Generate(letfHoldings)...)
	}
	return gatheredInsights, nil
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
		allHoldings = append(allHoldings, holdings...)
	}

	sort.Slice(allHoldings, func(i, j int) bool {
		return allHoldings[i].PercentContained > allHoldings[j].PercentContained
	})

	return allHoldings, nil
}

func main() {
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
	err = orchestrate(ctx, orchestrateRequest{
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
		insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator(overlap.Config{MinThreshold: config.Outputs.Insights.MinThresholdPercentage})},
		insightsLogger:    insights.NewInsightsLogger(insights.Config{RootDir: config.Directories.Artifacts + "/insights"}),
		websiteGenerators: []letf.Generator{letf.New(letf.Config{WebsiteDirectoryRoot: config.Directories.Websites, MinThreshold: config.Outputs.Websites.MinThresholdPercentage})},
	})
	if err != nil {
		log.Fatal(err)
	}
}
