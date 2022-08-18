package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// DownloadFile will download an url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	PanicErr(err)
	defer func(Body io.ReadCloser) {
		PanicErr(Body.Close())
	}(resp.Body)

	// Create the file
	out, err := os.Create(filepath)
	PanicErr(err)
	defer func(out *os.File) {
		PanicErr(out.Close())
	}(out)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

var replacements = map[string]string{
	"Triumph Group, Inc.,": "\"Triumph Group, Inc.\",",
}

func ReadCSVFromUrlWithLocalMasks(url string, temporaryDir string, comma rune, fieldsPerRecord int) ([][]string, error) {
	fileName := "/temp.csv"
	filePath := temporaryDir + fileName
	PanicErr(DownloadFile(filePath, url))

	input, err := ioutil.ReadFile(filePath)
	PanicErr(err)

	output := input
	for s, s2 := range replacements {
		output = bytes.Replace(output, []byte(s), []byte(s2), -1)
	}

	records := replaceStrings(temporaryDir, err, output, comma, fieldsPerRecord)

	return records, err
}

func replaceStrings(temporaryDir string, err error, output []byte, comma rune, fieldsPerRecord int) [][]string {
	modifiedFileName := "/temp_modified.csv"
	modifiedFilePath := temporaryDir + modifiedFileName
	if err = ioutil.WriteFile(modifiedFilePath, output, 0666); err != nil {
		PanicErr(err)
	}
	f, err := os.Open(modifiedFilePath)
	PanicErr(err)
	defer func(f *os.File) {
		PanicErr(f.Close())
	}(f)

	csvReader := csv.NewReader(f)
	csvReader.Comma = comma
	csvReader.FieldsPerRecord = fieldsPerRecord
	records, err := csvReader.ReadAll()
	PanicErr(err)
	return records
}

func ReadCSVFromUrl(url string, comma rune, fieldsPerRecord int) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "fetching %+v returned err", url)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	reader := csv.NewReader(resp.Body)
	reader.Comma = comma
	reader.FieldsPerRecord = fieldsPerRecord
	data, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading data for %v returned err: %w", url, err)
	}

	return data, nil
}

func ReadCSVFromLocalFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	PanicErr(err)
	defer func(f *os.File) {
		PanicErr(f.Close())
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	PanicErr(err)

	return records, nil
}
