package main

import "stocks/models"

var TopHardcodedETFs = []models.LETFAccountTicker{
	"TQQQ",
	"XLU",
	"SOXL",
	"SPY",
	"LABU",
	"VXUS",
	"EEM",
	"XLF",
	"FXI",
	"UPRO",
}

var TopHardcodedETFsMap = map[models.LETFAccountTicker]bool{
	"TQQQ": true,
	"XLU":  true,
	"SOXL": true,
	"SPY":  true,
	"LABU": true,
	"VXUS": true,
	"EEM":  true,
	"XLF":  true,
	"FXI":  true,
	"UPRO": true,
}
