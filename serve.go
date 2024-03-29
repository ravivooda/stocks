package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"stocks/database/insights"
	"stocks/external/stocks/alphavantage"
	"stocks/insights/overlap"
	"stocks/orchestrate"
	"stocks/utils"
	"stocks/website"
	"time"
)

func serve(
	config orchestrate.Config,
	ctx context.Context,
	insightsConfig insights.Config,
	generator overlap.Generator,
	logger insights.Logger,
	client alphavantage.Client,
	websitePaths website.Paths,
	fileAddr string,
) {
	file, err := ioutil.ReadFile(fileAddr)
	utils.PanicErr(err)

	metadata := website.Metadata{}

	utils.PanicErr(json.Unmarshal(file, &metadata))
	testDuration := time.Duration(int64(config.Secrets.TestConfig.MaxServerRunTime)) * time.Second

	// TODO: Improve on the hack below
	metadata.TemplateCustomMetadata.WebsitePaths = websitePaths

	beginServing(ctx, logger, generator, metadata, testDuration, client, website.Config{
		InsightsConfig: insightsConfig,
	})
}

func beginServing(
	ctx context.Context,
	logger insights.Logger,
	generator overlap.Generator,
	metadata website.Metadata,
	testDuration time.Duration,
	client alphavantage.Client,
	websiteConfig website.Config,
) {
	server := website.New(websiteConfig, website.Dependencies{
		Logger:       logger,
		Generator:    generator,
		AlphaVantage: client,
	}, metadata)
	fmt.Println("started serving!!")
	utils.PanicErr(server.StartServing(ctx, testDuration))
	fmt.Println("done serving")
}
