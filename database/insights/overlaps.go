package insights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"sort"
	"stocks/models"
	"stocks/utils"
)

type OverlapWrapper struct {
	Leverage string
	Analysis models.LETFOverlapAnalysis
}

type allOverlapsWrapper struct {
	All []OverlapWrapper
}

func (l *logger) LogOverlapAnalysisForHolder(lhs models.LETFAccountTicker, wrappers []OverlapWrapper) (FileName, error) {
	_, err := utils.MakeDirs([]string{l.overlapsDirectory(string(lhs))})
	utils.PanicErr(err)
	fileName, fileAddr := l.allOverlapsPercentageFilePath(string(lhs))
	b, err := json.Marshal(allOverlapsWrapper{All: wrappers})
	bytes.Replace(b, []byte(":NaN"), []byte(":null"), -1)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
	return FileName(fileName), nil
}

func (l *logger) overlapsFilePath(etfHolder string, etfHoldees string) (fileName string, fileAddr string) {
	fileName = fmt.Sprintf("%s_%s.json", etfHolder, etfHoldees)
	fileAddr = fmt.Sprintf("%s/%s", l.overlapsDirectory(etfHolder), fileName)
	return fileName, fileAddr
}

const allOverlapsFileName = "all.json"

func (l *logger) allOverlapsPercentageFilePath(etfHolder string) (fileName string, fileAddr string) {
	fileAddr = fmt.Sprintf("%s/%s", l.overlapsDirectory(etfHolder), allOverlapsFileName)
	return allOverlapsFileName, fileAddr
}

func (l *logger) overlapsDirectory(etfHolder string) (directory string) {
	directory = fmt.Sprintf("%s/%s", l.c.OverlapsDirectory, etfHolder)
	return directory
}

func (l *logger) FetchOverlapsWithoutDetailedOverlaps(etfName string) (map[string][]models.LETFOverlapAnalysis, error) {
	_, allOverlapsPercentageFilePath := l.allOverlapsPercentageFilePath(string(utils.FetchAccountTicker(etfName)))

	file, err := ioutil.ReadFile(allOverlapsPercentageFilePath)
	if err != nil {
		utils.LogErr(err)
		return map[string][]models.LETFOverlapAnalysis{}, nil
	}

	data := allOverlapsWrapper{}
	utils.PanicErr(json.Unmarshal(file, &data))

	var parsedOverlaps = map[string][]models.LETFOverlapAnalysis{}
	for _, wrapper := range data.All {
		var analyses = parsedOverlaps[wrapper.Leverage]
		if analyses == nil {
			analyses = []models.LETFOverlapAnalysis{}
		}
		analyses = append(analyses, wrapper.Analysis)
		parsedOverlaps[wrapper.Leverage] = analyses
	}

	return utils.SortOverlapsWithinLeverage(parsedOverlaps), nil
}

func (l *logger) fetchOverlap(fileAddr string) OverlapWrapper {
	file, err := ioutil.ReadFile(fileAddr)
	utils.PanicErr(err)

	data := OverlapWrapper{}
	utils.PanicErr(json.Unmarshal(file, &data))
	return data
}

func (l *logger) FetchOverlapDetails(lhs string, rhs []string) (models.LETFOverlapAnalysis, error) {
	lhsETFHoldings, _, err := l.FetchHoldings(lhs)
	utils.PanicErr(err)

	var rhsETFHoldings []models.LETFHolding
	for _, rh := range rhs {
		h, _, err := l.FetchHoldings(rh)
		utils.PanicErr(err)
		rhsETFHoldings = append(rhsETFHoldings, h...)
	}

	totalOverlapPercentage, details := l.g.Compare(
		utils.MapLETFHoldingsWithStockTicker(lhsETFHoldings),
		utils.MapLETFHoldingsWithStockTicker(rhsETFHoldings),
	)

	sort.Slice(details, func(i, j int) bool {
		return details[i].Percentage > details[j].Percentage
	})

	return models.LETFOverlapAnalysis{
		LETFHolder:        utils.FetchAccountTicker(lhs),
		LETFHoldees:       utils.FetchAccountTickers(rhs),
		OverlapPercentage: totalOverlapPercentage,
		DetailedOverlap:   &details,
	}, nil
}
