package insights

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"stocks/models"
	"stocks/utils"
)

type stockWrapper struct {
	Holdings []models.LETFHolding
}

func (l *logger) LogStocks(ctx context.Context, holdingsWithStockTickerMap map[models.StockTicker][]models.LETFHolding) ([]FileName, error) {
	var filesCreated []FileName
	for ticker, holdings := range holdingsWithStockTickerMap {
		fileName, err := l.logStock(ctx, ticker, holdings)
		utils.PanicErr(err)
		filesCreated = append(filesCreated, fileName)
	}
	return filesCreated, nil
}

func (l *logger) logStock(_ context.Context, ticker models.StockTicker, holdings []models.LETFHolding) (FileName, error) {
	fileName, fileAddr := l.stocksFilePaths(string(ticker))
	b, err := json.Marshal(stockWrapper{Holdings: holdings})
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
	return FileName(fileName), nil
}

func (l *logger) stocksFilePaths(stockName string) (string, string) {
	fileName := fmt.Sprintf("%s.json", stockName)
	fileAddr := fmt.Sprintf("%s/%s", l.c.StocksDirectory, fileName)
	return fileName, fileAddr
}

func (l *logger) FetchStock(stock string) ([]models.LETFHolding, error) {
	_, fileAddr := l.stocksFilePaths(stock)
	file, err := ioutil.ReadFile(fileAddr)
	if err != nil {
		utils.LogErr(err)
		return nil, errors.New(fmt.Sprintf("Sorry, we currently do not support ticker: %s", stock))
	}
	data := stockWrapper{}

	err = json.Unmarshal(file, &data)
	if err != nil {
		utils.LogErr(err)
		return nil, errors.New(fmt.Sprintf("Sorry, we currently do not support ticker: %s", stock))
	}
	return data.Holdings, nil
}

func (l *logger) HasStock(stock string) (bool, error) {
	_, fileAddr := l.stocksFilePaths(stock)
	_, err := ioutil.ReadFile(fileAddr)
	if err == nil {
		return true, nil
	}
	return false, err
}
