package insights

import (
	"context"
	"encoding/json"
	"fmt"
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
	utils.PanicErr(err)

	data := stockWrapper{}

	utils.PanicErr(json.Unmarshal(file, &data))
	return data.Holdings, nil
}
