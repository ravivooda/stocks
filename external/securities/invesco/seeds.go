package invesco

import (
	"context"
	"fmt"
	"stocks/models"
)

func (c *client) ListSeeds(context.Context) (seeds []models.Seed, err error) {
	seeds = []models.Seed{}
	for _, knownETF := range knownETFs {
		seeds = append(seeds, models.Seed{
			URL:    fmt.Sprintf(c.config.TickerFormattedURL, knownETF),
			Ticker: knownETF,
			Header: models.Header{
				SkippableLines:    1,
				ExpectedColumns:   nil,
				OutstandingShares: models.OutstandingShares{},
			},
			Provider: models.Invesco,
		})
	}
	return seeds, nil
}
