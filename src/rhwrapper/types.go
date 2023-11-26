package rhwrapper

import (
	"time"
)

type HistoricalData struct {
	Date        string
	Label       string
	Numerator   int
	Denominator int
}

type StockData struct {
	Symbol     string
	Historical []HistoricalData
	SplitMap   map[string]interface{} // You may need to specify the type more accurately based on your use case
}

type Profit struct {
	Date   string
	Amount float64
	Lcap   bool
	Ticker string
	Tag    string
}

type Gains struct {
	LcapAmount float64
	ScapAmount float64
}

type Stock struct {
	Action            string
	Qty               float64
	UnitCost          float64
	Datetime          time.Time
	Description       string
	OptionStrikePrice float64
	OptionExpiryDate  time.Time
}
