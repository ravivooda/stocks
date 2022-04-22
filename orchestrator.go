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
	"stocks/utils"
)

type orchestrateRequest struct {
	config            Config
	db                database.DB
	clients           map[models.Provider]securities.Client
	parsers           []alerts.AlertParser
	notifier          notifications.Notifier
	insightGenerators []overlap.Generator
	insightsLogger    insights.Logger
}

func orchestrate(ctx context.Context, request orchestrateRequest) error {
	seeds, err := request.db.ListSeeds(ctx)
	if err != nil {
		return err
	}

	holdings, err := fetchHoldings(ctx, seeds, request.clients)
	if err != nil {
		return err
	}

	holdingsMap := utils.MapLETFHoldingsWithStockTicker(holdings)
	//fmt.Println(holdingsMap)

	gatheredAlerts, err := gatherAlerts(ctx, request.parsers, holdingsMap)
	if err != nil {
		return err
	}
	fmt.Printf("Found alerts: %d\n", len(gatheredAlerts))

	if request.config.Notifications.ShouldSendEmails {
		_, err := request.notifier.SendAll(ctx, gatheredAlerts)
		if err != nil {
			return err
		}
	}

	gatheredInsights, err := gatherInsights(ctx, request.insightGenerators, holdings)
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

	return nil
}

func gatherInsights(
	_ context.Context,
	generators []overlap.Generator,
	letfHoldings []models.LETFHolding,
) ([]models.LETFOverlapAnalysis, error) {
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
		return allHoldings[i].Percent > allHoldings[j].Percent
	})

	return allHoldings, nil
}

func main() {
	ctx := context.Background()
	db := database.NewDumbDatabase()
	direxionClient, err := direxion.NewClient()
	microsectorClient, err := microsector.NewClient()
	if err != nil {
		log.Fatal(err)
		return
	}

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
	err = orchestrate(ctx, orchestrateRequest{
		config: config,
		db:     db,
		clients: map[models.Provider]securities.Client{
			models.Direxion:    direxionClient,
			models.MicroSector: microsectorClient,
		},
		parsers:           alertParsers,
		notifier:          notifier,
		insightGenerators: []overlap.Generator{overlap.NewOverlapGenerator()},
		insightsLogger:    insights.NewInsightsLogger(insights.Config{RootDir: config.Directories.Artifacts + "/insights"}),
	})
	if err != nil {
		log.Fatal(err)
	}
}
