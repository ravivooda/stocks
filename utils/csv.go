package utils

import (
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
)

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
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records, nil
}
