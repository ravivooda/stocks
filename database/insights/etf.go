package insights

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"stocks/models"
	"stocks/utils"
)

type holdingsWrapper struct {
	Leverage string
	Holdings []models.LETFHolding
}

func (l *logger) LogHoldings(_ context.Context, etfName models.LETFAccountTicker, holdings []models.LETFHolding, leverage string) (FileName, error) {
	fileName, fileAddr := l.holdingsFilePaths(string(etfName))
	b, err := json.Marshal(holdingsWrapper{
		Leverage: leverage,
		Holdings: holdings,
	})
	bytes.Replace(b, []byte(":NaN"), []byte(":null"), -1)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
	return FileName(fileName), nil
}

func (l *logger) holdingsFilePaths(etfName string) (string, string) {
	fileName := fmt.Sprintf("%s.json", etfName)
	fileAddr := fmt.Sprintf("%s/%s", l.c.ETFHoldingsDirectory, fileName)
	return fileName, fileAddr
}

func (l *logger) FetchHoldings(etfName string) ([]models.LETFHolding, string, error) {
	_, fileAddr := l.holdingsFilePaths(etfName)
	file, err := ioutil.ReadFile(fileAddr)
	if err != nil {
		utils.LogErr(err)
		return nil, "", errors.New(fmt.Sprintf("Sorry, cannot find ETF: %s", etfName))
	}

	data := holdingsWrapper{}

	utils.PanicErr(json.Unmarshal(file, &data))
	return data.Holdings, data.Leverage, nil
}
