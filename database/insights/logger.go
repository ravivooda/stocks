package insights

import (
	"encoding/csv"
	"fmt"
	"os"
	"stocks/models"
	"stocks/utils"
)

type FileName string

type Logger interface {
	Log(analysis models.LETFOverlapAnalysis) (FileName, error)
}

type Config struct {
	RootDir string
}

type logger struct {
	c Config
}

func (l *logger) Log(analysis models.LETFOverlapAnalysis) (FileName, error) {
	_, err := utils.MakeDirs([]string{l.c.RootDir})
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s_%s.csv", analysis.LETFHolding1, analysis.LETFHolding2)
	fileAddr := fmt.Sprintf("%s/%s", l.c.RootDir, fileName)
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
	err = csvWriter.Write([]string{"Stock Ticker", string(analysis.LETFHolding1), string(analysis.LETFHolding2), "Minimum"})
	if err != nil {
		return "", err
	}

	// Write detailed
	var (
		lsum = float64(0)
		rsum = float64(0)
	)

	for _, overlap := range analysis.DetailedOverlap {
		lPercent := overlap.IndividualPercentagesMap[analysis.LETFHolding1]
		rPercent := overlap.IndividualPercentagesMap[analysis.LETFHolding2]
		err = csvWriter.Write([]string{string(overlap.Ticker), floatToString(lPercent), floatToString(rPercent), floatToString(overlap.Percentage)})
		if err != nil {
			return "", err
		}
		lsum += lPercent
		rsum += rPercent
	}

	// Last summation row
	err = csvWriter.Write([]string{"", floatToString(lsum), floatToString(rsum), floatToString(analysis.OverlapPercentage)})
	if err != nil {
		return "", err
	}

	return FileName(fileName), nil
}

func floatToString(input float64) string {
	return fmt.Sprintf("%.2f", input)
}

func NewInsightsLogger(config Config) Logger {
	return &logger{c: config}
}
