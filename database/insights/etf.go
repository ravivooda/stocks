package insights

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"stocks/models"
	"stocks/utils"
)

type holdingsWrapper struct {
	Holdings []models.LETFHolding
}

func (l *logger) LogHoldings(
	_ context.Context,
	etfName models.LETFAccountTicker,
	holdings []models.LETFHolding,
	leverageMappedOverlaps map[string][]models.LETFOverlapAnalysis,
) (FileName, error) {
	fileName, fileAddr := l.filePaths(string(etfName))
	b, err := json.Marshal(holdingsWrapper{Holdings: holdings})
	bytes.Replace(b, []byte(":NaN"), []byte(":null"), -1)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
	return FileName(fileName), nil
}

func (l *logger) filePaths(etfName string) (string, string) {
	fileName := fmt.Sprintf("%s.json", etfName)
	fileAddr := fmt.Sprintf("%s/%s", l.c.ETFHoldingsDirectory, fileName)
	return fileName, fileAddr
}

func (l *logger) FetchHoldings(etfName string) ([]models.LETFHolding, error) {
	_, fileAddr := l.filePaths(etfName)
	file, err := ioutil.ReadFile(fileAddr)
	utils.PanicErr(err)

	data := holdingsWrapper{}

	utils.PanicErr(json.Unmarshal(file, &data))
	return data.Holdings, nil
}
