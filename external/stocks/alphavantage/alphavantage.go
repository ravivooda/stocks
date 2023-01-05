package alphavantage

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const formattedURL = "https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=compact&apikey=%s"

type Client struct {
	cache  map[string]Response
	config Config
}

func (c Client) FetchStockTradingData(ticker string) (Response, error) {
	if resp, ok := c.cache[ticker]; ok {
		return resp, nil
	}

	resp, err := c.loadFromRemote(ticker)
	if err != nil {
		return Response{}, err
	}
	c.cache[ticker] = resp
	return c.cache[ticker], nil
}

func (c Client) loadFromRemote(ticker string) (Response, error) {
	url := fmt.Sprintf(formattedURL, ticker, c.config.APIKey)
	//fmt.Printf("[Alpha Vantage] Making request: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Response{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}

	resp := Response{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return Response{}, err
	}
	return resp, nil
}

type Config struct {
	FormattedURL string
	APIKey       string `mapstructure:"key"`
	Retries      int
}

func New(config Config) Client {
	return Client{
		cache:  map[string]Response{},
		config: config,
	}
}
