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

type overlapWrapper struct {
	Leverage string
	Analysis models.LETFOverlapAnalysis
}

func (l *logger) LogOverlapAnalysis(leverage string, analysis models.LETFOverlapAnalysis) (FileName, error) {
	_, err := utils.MakeDirs([]string{l.overlapsDirectory(analysis.LETFHolder)})
	utils.PanicErr(err)
	fileName, fileAddr := l.overlapsFilePath(analysis.LETFHolder, analysis.LETFHoldees)
	b, err := json.Marshal(overlapWrapper{
		Leverage: leverage,
		Analysis: analysis,
	})
	bytes.Replace(b, []byte(":NaN"), []byte(":null"), -1)
	utils.PanicErr(err)

	utils.PanicErr(ioutil.WriteFile(fileAddr, b, fs.ModePerm))
	return FileName(fileName), nil
}

func (l *logger) overlapsFilePath(etfHolder models.LETFAccountTicker, etfHoldees []models.LETFAccountTicker) (fileName string, fileAddr string) {
	fileName = fmt.Sprintf("%s_%s.json", etfHolder, utils.JoinLETFAccountTicker(etfHoldees, "_"))
	fileAddr = fmt.Sprintf("%s/%s", l.overlapsDirectory(etfHolder), fileName)
	return fileName, fileAddr
}

func (l *logger) overlapsDirectory(etfHolder models.LETFAccountTicker) (directory string) {
	directory = fmt.Sprintf("%s/%s", l.c.OverlapsDirectory, etfHolder)
	return directory
}

func (l *logger) FetchOverlaps(etfName string) (map[string][]models.LETFOverlapAnalysis, error) {
	directory := l.overlapsDirectory(utils.FetchAccountTicker(etfName))
	fileInfos, err := ioutil.ReadDir(directory)
	utils.PanicErr(err)

	var parsedOverlaps = map[string][]models.LETFOverlapAnalysis{}
	for _, info := range fileInfos {
		file, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", directory, info.Name()))
		utils.PanicErr(err)

		data := overlapWrapper{}
		utils.PanicErr(json.Unmarshal(file, &data))
		var analyses = parsedOverlaps[data.Leverage]
		if analyses == nil {
			analyses = []models.LETFOverlapAnalysis{}
		}
		analyses = append(analyses, data.Analysis)
		parsedOverlaps[data.Leverage] = analyses
	}

	var sortedOverlaps = map[string][]models.LETFOverlapAnalysis{}
	for s, analyses := range parsedOverlaps {
		sort.Slice(analyses, func(i, j int) bool {
			return analyses[i].OverlapPercentage > analyses[j].OverlapPercentage
		})
		sortedOverlaps[s] = analyses
	}

	return sortedOverlaps, nil
}
