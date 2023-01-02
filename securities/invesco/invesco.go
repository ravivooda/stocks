package invesco

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"stocks/models"
	"stocks/securities"
	"stocks/securities/invesco/template1"
	"stocks/securities/invesco/template2"
	"stocks/securities/invesco/template3"
	"stocks/securities/invesco/template4"
	"stocks/utils"
	"strings"
)

type client struct {
	config Config
}

func (c *client) GetHoldings(_ context.Context, seed models.Seed, etf models.ETF) ([]models.LETFHolding, error) {
	data, err := utils.ReadCSVFromUrl(seed.URL, ',', -1)
	utils.PanicErr(err)

	if len(data) <= seed.Header.SkippableLines {
		utils.PanicErr(errors.New(fmt.Sprintf("got fewer (%d) than expected lines (%d) for seed %+v", len(data), seed.Header.SkippableLines, seed)))
		//fmt.Printf("invesco error: got fewer (%d) than expected lines (%d) for seed %+v", len(data), seed.Header.SkippableLines, seed)
		return nil, err
	}

	observedHeaders := strings.Join(utils.Trimmed(data[seed.Header.SkippableLines-1]), ",")
	if observedHeaders == strings.Join(template1.Headers, ",") {
		return template1.Parse(data, seed, etf)
	} else if observedHeaders == strings.Join(template2.Headers, ",") {
		return template2.Parse(data, seed, etf)
	} else if observedHeaders == strings.Join(template3.Headers, ",") {
		return template3.Parse(data, seed, etf)
	} else if observedHeaders == strings.Join(template4.Headers, ",") {
		return template4.Parse(data, seed, etf)
	}

	panic(fmt.Sprintf("invesco error: columns did not match -> expected: (%s), received: (%s) for seed %+v\n", seed.Header.ExpectedColumns, data[seed.Header.SkippableLines-1], seed))
	return nil, err
}

type Config struct {
	TickerFormattedURL string `mapstructure:"url"`
}

func New(config Config) securities.SeedProvider {
	return &client{
		config: config,
	}
}
