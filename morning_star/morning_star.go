package morning_star

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"stocks/models"
)

type MSAPI interface {
	GetMovers(ctx context.Context) (models.MSResponse, error)
}

type msapi struct {
	config Config
}

func (m *msapi) GetMovers(_ context.Context) (models.MSResponse, error) {
	req, err := http.NewRequest("GET", m.config.URL, nil)
	if err != nil {
		return models.MSResponse{}, err
	}

	req.Header.Add("X-RapidAPI-Host", m.config.Host)
	req.Header.Add("X-RapidAPI-Key", m.config.Key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.MSResponse{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return models.MSResponse{}, err
	}

	var result models.MSResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
		return models.MSResponse{}, err
	}

	return result, nil
}

type Config struct {
	URL  string
	Host string
	Key  string
}

func New(config Config) MSAPI {
	return &msapi{config: config}
}
