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
	_, err := utils.MakeDirs([]string{l.overlapsDirectory(string(analysis.LETFHolder))})
	utils.PanicErr(err)
	fileName, fileAddr := l.overlapsFilePath(string(analysis.LETFHolder), utils.JoinLETFAccountTicker(analysis.LETFHoldees, "_"))
	b, err := json.Marshal(overlapWrapper{
		Leverage: leverage,
		Analysis: analysis,
	})
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

func (l *logger) overlapsDirectory(etfHolder string) (directory string) {
	directory = fmt.Sprintf("%s/%s", l.c.OverlapsDirectory, etfHolder)
	return directory
}

func (l *logger) FetchOverlaps(etfName string) (map[string][]models.LETFOverlapAnalysis, error) {
	directory := l.overlapsDirectory(string(utils.FetchAccountTicker(etfName)))
	fileInfos, err := ioutil.ReadDir(directory)
	utils.PanicErr(err)

	var parsedOverlaps = map[string][]models.LETFOverlapAnalysis{}
	for _, info := range fileInfos {
		data := l.fetchOverlap(fmt.Sprintf("%s/%s", directory, info.Name()))
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

func (l *logger) fetchOverlap(fileAddr string) overlapWrapper {
	file, err := ioutil.ReadFile(fileAddr)
	utils.PanicErr(err)

	data := overlapWrapper{}
	utils.PanicErr(json.Unmarshal(file, &data))
	return data
}

func (l *logger) FetchOverlap(holder string, holdees string) (models.LETFOverlapAnalysis, error) {
	_, fileAddr := l.overlapsFilePath(holder, holdees)
	o := l.fetchOverlap(fileAddr)
	sort.Slice(*o.Analysis.DetailedOverlap, func(i, j int) bool {
		return (*o.Analysis.DetailedOverlap)[i].Percentage > (*o.Analysis.DetailedOverlap)[j].Percentage
	})
	return o.Analysis, nil
}
