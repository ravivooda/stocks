package alphavantage

import "sort"

type DailyData struct {
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

type Response struct {
	MetaData struct {
		Information   string `json:"1. Information"`
		Symbol        string `json:"2. Symbol"`
		LastRefreshed string `json:"3. Last Refreshed"`
		OutputSize    string `json:"4. Output Size"`
		TimeZone      string `json:"5. Time Zone"`
	} `json:"Meta Data"`
	TimeSeriesDaily map[string]DailyData `json:"Time Series (Daily)"`
}

func (r Response) LatestDate() string {
	retDate := ""
	for date := range r.TimeSeriesDaily {
		if date > retDate {
			retDate = date
		}
	}
	return retDate
}

type LinearTimeSeriesDaily struct {
	Date       string
	DailyPrice string
}

func (r Response) SplitByXDates(chunkSize int) (rets []LinearTimeSeriesDaily) {
	items := r.sorted()
	if len(items) <= chunkSize {
		return items
	}

	lastDate := items[len(items)-1]

	for chunkSize < len(items) {
		rets, items = append(rets, items[0]), items[chunkSize:]
	}

	rets = append(rets, lastDate)
	return
}

func (r Response) sorted() []LinearTimeSeriesDaily {
	var rets []LinearTimeSeriesDaily
	for date, data := range r.TimeSeriesDaily {
		rets = append(rets, LinearTimeSeriesDaily{
			Date:       date,
			DailyPrice: data.AdjustedClose,
		})
	}

	sort.Slice(rets, func(i, j int) bool {
		return rets[i].Date < rets[j].Date
	})
	return rets
}
