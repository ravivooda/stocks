package database

import (
	"context"
	"fmt"
	"stocks/models"
)

type dumbDatabase struct {
}

func (i dumbDatabase) ListSeeds(_ context.Context) ([]models.Seed, error) {
	expectedColumns := []string{
		"TradeDate",
		"AccountTicker",
		"StockTicker",
		"SecurityDescription",
		"Shares",
		"Price",
		"MarketValue",
	}
	header := models.Header{
		SkippableLines:  4,
		ExpectedColumns: expectedColumns,
		OutstandingShares: models.OutstandingShares{
			LineNumber: 2,
			Prefix:     "Shares Outstanding:",
		},
	}
	var stocks = []string{
		"DFEN",
		"WANT",
		"WEBS",
		"WEBL",
		"FAZ",
		"FAS",
		"CURE",
		"NAIL",
		"DUSL",
		"PILL",
		"DRV",
		"DRN",
		"DPST",
		"RETL",
		"HIBS",
		"HIBL",
		"LABD",
		"LABU",
		"SOXS",
		"SOXL",
		"TECS",
		"TECL",
		"TPOR",
		"UTSL",
		"TMV",
		"TMF",
		"TYO",
		"TYD",
		"MIDU",
		"SPXS",
		"SPXL",
		"TZA",
		"TNA",
		"YANG",
		"YINN",
		"EURL",
		"EDZ",
		"EDC",
		"MEXX",
		"KORU",
		"TENG",
		"CHAU",
		"CWEB",
		"CLDS",
		"CLDL",
		"ERY",
		"ERX",
		"FNTC",
		"DUST",
		"NUGT",
		"JDST",
		"JNUG",
		"BRZU",
		"INDL",
		"MNM",
		"ONG",
		"UBOT",
		"RUSL",
		"SPUU",
		"EVEN",
		"DRIP",
		"GUSH",
		"FNGG",
		"SWAR",
		"OOTO",
		"DOZR",
		"KLNE",
		"CHAD",
		"SPDN",
	}
	var rets []models.Seed
	for _, stock := range stocks {
		rets = append(rets, models.Seed{
			URL:    fmt.Sprintf("https://www.direxion.com/holdings/%s.csv", stock),
			Ticker: stock,
			Header: header,
		})
	}
	var weirdStocks = []string{
		"COM",
		"NIFE",
		"LOPX",
		"QQQE",
		"RWGV",
		"RWVG",
		"HJEN",
		"MOON",
		"TYNE",
		"WFH",
		"WWOW",
		"MSGR",
	}
	for _, weirdStock := range weirdStocks {
		rets = append(rets, models.Seed{
			URL:    fmt.Sprintf("https://www.direxion.com/holdings/%s.csv", weirdStock),
			Ticker: weirdStock,
			Header: models.Header{
				SkippableLines: 4,
				ExpectedColumns: []string{
					"TradeDate",
					"AccountTicker",
					"StockTicker",
					"SecurityDescription",
					"Shares",
					"Price",
					"MarketValue",
					"Cusip",
					"HoldingsPercent",
				},
				OutstandingShares: models.OutstandingShares{
					LineNumber: 2,
					Prefix:     "Shares Outstanding:",
				},
			},
		})
	}
	rets = append(rets)
	return rets, nil
}

func NewDumbDatabase() DB {
	return &dumbDatabase{}
}
