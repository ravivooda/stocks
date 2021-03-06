package proshares

import "stocks/models"

var (
	ignoreHoldings = map[models.LETFAccountTicker]bool{
		"SPXB": true,
		"TBF":  true,
		"DXD":  true,
		"EFZ":  true,
		"QID":  true,
		"SPXU": true,
		"SDOW": true,
		"SEF":  true,
		"REW":  true,
		"SSG":  true,
		"UBT":  true,
		"HYHG": true,
		"SBB":  true,
		"UPV":  true,
		"ULE":  true,
		"DDG":  true,
		"SCO":  true,
		"URTY": true,
		"GLL":  true,
		"BOIL": true,
	}
	knowinglyIgnoredIssues = map[models.LETFAccountTicker]bool{
		// TODO: Short ETFs are ignored, but should be handled differently
		"EWV":  true,
		"AGQ":  true,
		"BIS":  true,
		"BITO": true,
		"BITI": true,
		"BZQ":  true,
		"CROC": true,
		"CSM":  true,
		"DOG":  true,
		"DUG":  true,
		"EET":  true,
		"EEV":  true,
		"EFO":  true,
		"EFU":  true,
		"EMSH": true,
		"EMTY": true,
		"EPV":  true,
		"EUFX": true,
		"EUM":  true,
		"EUO":  true,
		"FINU": true,
		"FINZ": true,
		"FUT":  true,
		"FXP":  true,
		"HDG":  true,
		"IGHG": true,
		"KOLD": true,
		"MYY":  true,
		"MZZ":  true,
		"OILD": true,
		"OILK": true,
		"OILU": true,
		"PSQ":  true,
		"PST":  true,
		"REK":  true,
		"RINF": true,
		"RWM":  true,
		"RXD":  true,
		"SBM":  true,
		"SCC":  true,
		"SCOM": true,
		"SDD":  true,
		"SDP":  true,
		"SDS":  true,
		"SH":   true,
		"SIJ":  true,
		"SJB":  true,
		"SKF":  true,
		"SMDD": true,
		"SMN":  true,
		"SPXT": true,
		"SQQQ": true,
		"SRS":  true,
		"SRTY": true,
		"SVXY": true,
		"SZK":  true,
		"TBT":  true,
		"TBX":  true,
		"TTT":  true,
		"TWM":  true,
		"UBIO": true,
		"UBR":  true,
		"UCO":  true,
		"UCOM": true,
		"UGL":  true,
		"UJB":  true,
		"UST":  true,
		"UVXY": true,
		"UWM":  true,
		"UXI":  true,
		"VIXM": true,
		"VIXY": true,
		"XCOM": true,
		"XPP":  true,
		"YCL":  true,
		"YCOM": true,
		"YCS":  true,
		"YXI":  true,
		"ZBIO": true,
		"ZSL":  true,
		// TODO: interim issues with csvs
		"ALTS": true,
	}
)
