package main

import (
	"context"
	"fmt"
	"log"
	"stocks/database"
	"stocks/database/etfdb"
	"stocks/database/insights"
	"stocks/insights/overlap"
	"stocks/models"
	"stocks/orchestrate"
	"stocks/utils"
	"stocks/website"
)

func main() {
	ctx, db, etfs, websitePaths := defaults()
	config, insightsConfig := createConfigs()
	metadataFileAddr := fmt.Sprintf("%s/%s.json", insightsConfig.RootDirectory, "metadata")
	autoCompleteMetadataFileAddr := fmt.Sprintf("%s/%s.json", insightsConfig.RootDirectory, "autocomplete_metadata")
	generators, logger := logicProviders(config, insightsConfig)
	if config.Secrets.Uploads.ShouldUploadInsightsOutputToGCP {
		setup(ctx, false, setupRequest{
			metadataFileDestination:     metadataFileAddr,
			autoCompleteFileDestination: autoCompleteMetadataFileAddr,
			config:                      config,
			insightsConfig:              insightsConfig,
			db:                          db,
			etfs:                        etfs,
			generators:                  generators,
			logger:                      logger,
		})
	} else {
		// TODO: Hardcoded 0 index lookup on generators[0] in the line below
		serve(config, ctx, insightsConfig, generators[0], logger, websitePaths, metadataFileAddr)
	}
}

func logicProviders(config orchestrate.Config, insightsConfig insights.Config) ([]overlap.Generator, insights.Logger) {
	generator := overlap.NewOverlapGenerator(config.Outputs.Insights)
	return []overlap.Generator{generator}, insights.NewInsightsLogger(insightsConfig, generator)
}

func defaults() (context.Context, database.DB, []models.ETF, website.Paths) {
	ctx := context.Background()
	db := database.NewDumbDatabase()

	etfsGenerator := etfdb.New(etfdb.Config{})
	etfs, err := etfsGenerator.ListETFs(ctx)
	utils.PanicErr(err)
	fmt.Printf("Found %d etfs\n", len(etfs))

	websitePaths := website.DefaultWebsitePaths
	return ctx, db, etfs, websitePaths
}

func createConfigs() (orchestrate.Config, insights.Config) {
	config, err := orchestrate.NewConfig()
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
	return config, insightsConfig
}
