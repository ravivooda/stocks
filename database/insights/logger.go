package insights

import (
	"encoding/csv"
	"fmt"
	"os"
	"stocks/models"
	"stocks/utils"
)

type logger struct {
	c Config
}

func (l *logger) LogOverlapAnalysis(analysis models.LETFOverlapAnalysis) (FileName, error) {
	_, err := utils.MakeDirs([]string{l.c.OverlapsDirectory})
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s_%s.csv", analysis.LETFHolder, utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"))
	fileAddr := fmt.Sprintf("%s/%s", l.c.OverlapsDirectory, fileName)
	csvFile, err := os.Create(fileAddr)
	defer func(csvFile *os.File) {
		_ = csvFile.Close()
	}(csvFile)
	if err != nil {
		return "", err
	}

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()
	// Heading
	err = csvWriter.Write([]string{"Stock Ticker", string(analysis.LETFHolder), string(analysis.LETFHoldees[0]), "Minimum"})
	// TODO: Fix the zero index assumption made above
	if err != nil {
		return "", err
	}

	// Write detailed
	var (
		lsum = float64(0)
		rsum = map[int]float64{}
	)

	for _, overlap := range *analysis.DetailedOverlap {
		lPercent := overlap.IndividualPercentagesMap[analysis.LETFHolder]
		// TODO: Fix the zero index assumption made above

		columnsToWrite := []string{string(overlap.Ticker), floatToString(lPercent)}
		for i, holdee := range analysis.LETFHoldees {
			rPercent := overlap.IndividualPercentagesMap[holdee]
			rPercentStrings := floatToString(rPercent)
			columnsToWrite = append(columnsToWrite, rPercentStrings)
			rsum[i] = rsum[i] + rPercent
		}
		columnsToWrite = append(columnsToWrite, floatToString(overlap.Percentage))
		err = csvWriter.Write(columnsToWrite)
		if err != nil {
			return "", err
		}
		lsum += lPercent
	}

	// Last summation row
	lastSummationRow := []string{"", floatToString(lsum)}
	lastSummationRow = append(lastSummationRow, orderExactly(rsum)...)
	lastSummationRow = append(lastSummationRow, floatToString(analysis.OverlapPercentage))
	err = csvWriter.Write(lastSummationRow)
	if err != nil {
		return "", err
	}

	return FileName(fileName), nil
}

func orderExactly(input map[int]float64) []string {
	var rets []string
	for i := 0; i < len(input); i++ {
		rets = append(rets, floatToString(input[i]))
	}
	return rets
}

func floatToString(input float64) string {
	return fmt.Sprintf("%.2f", input)
}

func NewInsightsLogger(config Config) Logger {
	_, err := utils.MakeDirs([]string{config.ETFHoldingsDirectory})
	utils.PanicErr(err)
	return &logger{c: config}
}
