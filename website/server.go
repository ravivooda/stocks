package website

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"stocks/database/insights"
	"stocks/models"
	"stocks/website/letf"
	"strings"
)

type Server interface {
}

type Config struct {
	InsightsConfig insights.Config
}

type server struct {
	config       Config
	WebsitePaths letf.WebsitePaths
}

func (s *server) StartServing(ctx context.Context, generator letf.Generator) {
	http.HandleFunc("/etf-summary", func(writer http.ResponseWriter, request *http.Request) {

		pathSplits := strings.Split(request.URL.Path, "/")
		if len(pathSplits) < 2 {
			log.Printf("couldn't parse uri %s\n", request.URL.Path)
			return
		}
		etfName := pathSplits[1]

		data := struct {
			AccountTicker models.LETFAccountTicker
			Holdings      []models.LETFHolding
			Overlaps      map[string][]models.LETFOverlapAnalysis
			AccountsMap   map[models.LETFAccountTicker][]models.LETFHolding
			WebsitePaths  letf.WebsitePaths
		}{
			AccountTicker: accountTicker,
			Holdings:      letfHoldings,
			Overlaps:      allAnalysis,
			AccountsMap:   letfs,
			WebsitePaths:  websitePaths,
		}

		_, err := generateHTML(letf.SummaryTemplateLoc, writer, data)
		if err != nil {
			log.Printf("couldn't generate response for uri %s\n", request.URL.Path)
			return
		}
	})
}

func generateHTML(templateLoc string, outputFilePath io.Writer, data interface{}) (bool, error) {
	t := template.Must(template.ParseFiles(templateLoc))
	err := t.Execute(outputFilePath, data)
	if err != nil {
		return false, err
	}
	return false, nil
}

func New(config Config) Server {
	return &server{config: config}
}
