package rhwrapper

// package for interacting with robinhood API

import (
	"context"
	"encoding/gob"
	"fmt"
	robinhood "github.com/Ryang20718/robinhood-client/client"
	models "github.com/Ryang20718/robinhood-client/models"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"os"
	"sort"
	"strings"
)

type HoodAPI interface {
	Auth(username string, password string, mfa string) error
	FetchOptionTrades(ctx context.Context) (map[string][]models.OptionTransaction, error)
	FetchRegularTrades(ctx context.Context) (map[string][]models.Transaction, error)
}

type Hood struct {
	Cli *robinhood.Client
}

var SymbolChangeCache = make(map[string]string) // mapping of original symbol --> current symbol

func (h *Hood) Auth(username string, password string, mfa string) (*robinhood.Client, error) {
	if username == "" {
		return nil, fmt.Errorf("requires a username")
	}
	if password == "" {
		return nil, fmt.Errorf("requires a password")
	}
	if mfa == "" {
		return nil, fmt.Errorf("requires an mfa")
	}
	cli, err := robinhood.Dial(
		&robinhood.OAuth{
			Username: username,
			Password: password,
			MFA:      mfa,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to auth rhood err: %v", err.Error())
	}
	return cli, nil
}

/*
Returns mapping of ticker to model.OptionTransaction

Each ticker maps to a list which is sorted by created datetime
*/
func (h *Hood) FetchOptionTrades(ctx context.Context) (map[string][]models.OptionTransaction, error) {
	optionsOrderMap := make(map[string][]models.OptionTransaction)
	optionOrders, err := h.Cli.GetOptionsOrders(ctx)
	if err != nil {
		return nil, err
	}
	if optionOrders == nil {
		return optionsOrderMap, nil
	}
	for _, order := range *optionOrders {
		if optionsOrderMap[order.Ticker] == nil {
			optionsOrderMap[order.Ticker] = []models.OptionTransaction{}
		}
		optionsOrderMap[order.Ticker] = append(optionsOrderMap[order.Ticker], order)
	}
	return optionsOrderMap, nil
}

/*
Returns mapping of ticker to model.Transaction

Each ticker maps to a list which is sorted by created datetime
*/
func (h *Hood) FetchRegularTrades(ctx context.Context) (map[string][]models.Transaction, error) {
	stockOrderMap := make(map[string][]models.Transaction)
	stockOrders, err := h.Cli.GetStockOrders()
	if err != nil {
		return nil, err
	}
	for _, order := range stockOrders {
		if stockOrderMap[order.Ticker] == nil {
			stockOrderMap[order.Ticker] = []models.Transaction{}
		}
		stockOrderMap[order.Ticker] = append(stockOrderMap[order.Ticker], order)
	}
	return stockOrderMap, nil
}

func (h *Hood) FetchStockSplits(ctx context.Context, ticker string) error {
	return nil
}

/*
Fetch current ticker symbol
*/
func (h *Hood) FetchCurrentTickerSymbol(symbol string) (string, error) {
	if _, symbolInCache := SymbolChangeCache[symbol]; symbolInCache {
		// cached value
		return SymbolChangeCache[symbol], nil
	}
	_, symbolFound := h.Cli.GetInstrumentForSymbol(symbol)
	if symbolFound != nil {
		// this only occurs if the symbol is no longer found
		newSymbol, err := FetchStockSymbolChange(symbol)
		if err != nil {
			return "", err
		}
		SymbolChangeCache[symbol] = newSymbol
	} else {
		SymbolChangeCache[symbol] = symbol
	}

	return SymbolChangeCache[symbol], nil
}

/*
convert profit to dataframe
*/
func (h *Hood) ConvertProfitDf(profitList []Profit) *dataframe.DataFrame {
	// Create series for each field
	years := series.New([]string{}, series.String, "Year")
	dates := series.New([]string{}, series.String, "Date")
	amounts := series.New([]float64{}, series.Float, "Amount")
	lcaps := series.New([]bool{}, series.Bool, "Lcap")
	tickers := series.New([]string{}, series.String, "Ticker")
	tags := series.New([]string{}, series.String, "Tag")

	// Populate series with data from Profit struct array
	for _, profit := range profitList {
		years.Append(strings.Split(profit.Date, "-")[0])
		dates.Append(profit.Date)
		amounts.Append(profit.Amount)
		lcaps.Append(profit.Lcap)
		tickers.Append(profit.Ticker)
		tags.Append(profit.Tag)
	}

	// Create DataFrame
	df := dataframe.New(
		years,
		dates,
		amounts,
		lcaps,
		tickers,
		tags,
	)

	return &df
}

func (h *Hood) ProcessRealizedEarnings(ctx context.Context) (*dataframe.DataFrame, error) {
	if len("GG") == 10 {
		stockMap, err := h.FetchRegularTrades(ctx)
		if err != nil {
			panic("GG")
		}

		optionMap, err := h.FetchOptionTrades(ctx)
		if err != nil {
			panic("GG")
		}

		encodeFile, err := os.Create("/Users/ryang/Documents/rh_metrics/stock.map")
		if err != nil {
			panic(err)
		}

		// Since this is a binary format large parts of it will be unreadable
		encoder := gob.NewEncoder(encodeFile)

		// Write to the file
		if err := encoder.Encode(stockMap); err != nil {
			panic(err)
		}
		encodeFile.Close()

		encodeFile, err = os.Create("/Users/ryang/Documents/rh_metrics/option.map")
		if err != nil {
			panic(err)
		}

		// Since this is a binary format large parts of it will be unreadable
		encoder = gob.NewEncoder(encodeFile)

		// Write to the file
		if err := encoder.Encode(optionMap); err != nil {
			panic(err)
		}
		encodeFile.Close()
		// panic("GG")
	}

	var stockMap map[string][]models.Transaction
	dataFile, err := os.Open("stock.map")

	if err != nil {
		return nil, err
	}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&stockMap)

	if err != nil {
		return nil, err
	}
	dataFile.Close()

	var optionMap map[string][]models.OptionTransaction
	dataFile, err = os.Open("option.map")

	if err != nil {
		return nil, err
	}

	dataDecoder = gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&optionMap)

	if err != nil {
		return nil, err
	}
	dataFile.Close()

	stockList := []models.Transaction{}
	optionList := []models.OptionTransaction{}

	for _, vals := range stockMap {
		for _, val := range vals {
			stockList = append(stockList, val)
		}
	}
	for _, vals := range optionMap {
		for _, val := range vals {
			optionList = append(optionList, val)
		}
	}
	sort.Slice(stockList, func(i, j int) bool {
		return stockList[i].CreatedAt < stockList[j].CreatedAt
	})
	sort.Slice(optionList, func(i, j int) bool {
		return optionList[i].CreatedAt < optionList[j].CreatedAt
	})

	stockLen := len(stockList)
	optionLen := len(optionList)

	stockIdx, optionIdx := 0, 0
	profitList := []Profit{}
	profitsMap := make(map[string][]*models.Transaction) // keep track of buy/sell
	for {
		// interweave stocks & options to ensure FIFO
		if stockIdx >= stockLen && optionIdx >= optionLen {
			break
		}

		calcOption := false
		if stockIdx >= stockLen {
			calcOption = true
		} else if optionIdx >= optionLen {
			calcOption = false
		} else {
			// option date is before stock
			if strings.Split(optionList[optionIdx].CreatedAt, "T")[0] < strings.Split(stockList[stockIdx].CreatedAt, " ")[0] {
				calcOption = true
			} else {
				calcOption = false
			}
		}
		if calcOption {
			option := optionList[optionIdx]
			optionIdx += 1
			optionTicker, err := h.FetchCurrentTickerSymbol(option.Ticker)
			if err != nil {
				return nil, err
			}
			if option.Status != "Expired" && option.Status != "Assigned" {
				continue
			} else if option.Status == "Assigned" {
				premium := 0.0
				if option.TransactionType == "STO" {
					premium = option.UnitCost * -1
				} else {
					premium = option.UnitCost
				}
				stock := models.Transaction{
					Ticker:          optionTicker,
					TransactionType: "buy",
					Qty:             100.00 * option.Qty,
					UnitCost:        option.StrikePrice + premium,
					CreatedAt:       option.ExpirationDate,
					Tag:             option.TransactionType + " assigned",
				}
				if profitsMap[optionTicker] == nil {
					profitsMap[optionTicker] = []*models.Transaction{}
				}
				profitsMap[optionTicker] = append(profitsMap[optionTicker], &stock)
			} else if option.Status == "Expired" {
				if option.TransactionType == "STO" || option.TransactionType == "STC" {
					profit := Profit{
						Date:   strings.Split(option.CreatedAt, "T")[0],
						Amount: option.Qty * 100 * option.UnitCost,
						Lcap:   false, // TODO calculate whether actual LTG
						Ticker: optionTicker,
						Tag:    option.Tag,
					}
					profitList = append(profitList, profit)
				} else {
					profit := Profit{
						Date:   strings.Split(option.CreatedAt, "T")[0],
						Amount: -option.Qty * 100 * option.UnitCost,
						Lcap:   false, // TODO calculate whether actual LTG
						Ticker: optionTicker,
						Tag:    option.Tag,
					}
					profitList = append(profitList, profit)
				}
			}

		} else {
			stock := stockList[stockIdx]
			stockIdx += 1
			stockTicker, err := h.FetchCurrentTickerSymbol(stock.Ticker)
			if err != nil {
				return nil, err
			}
			if stock.TransactionType == "sell" {
				qty := stock.Qty
				indexToPop := -1
				lcapGain := 0.0
				scapGain := 0.0
				for qty != 0.0 {
					if len(profitsMap[stockTicker]) <= 0 {
						profit := Profit{
							Date:   strings.Split(stock.CreatedAt, " ")[0],
							Amount: stock.UnitCost * stock.Qty,
							Lcap:   false,
							Ticker: stockTicker,
							Tag:    stock.Tag,
						}
						profitList = append(profitList, profit)
						break
					}
					for i, boughtStock := range profitsMap[stockTicker] {
						if profitsMap[stockTicker][i].Qty > qty {
							gain := qty * (stock.UnitCost - boughtStock.UnitCost)
							if OneYearApart(boughtStock.CreatedAt, stock.CreatedAt) {
								lcapGain += gain
							} else {
								scapGain += gain
							}
							profitsMap[stockTicker][i].Qty -= qty
							qty = 0
							break
						} else {
							qty -= boughtStock.Qty
							indexToPop = i
							gain := boughtStock.Qty * (stock.UnitCost - boughtStock.UnitCost)
							if OneYearApart(boughtStock.CreatedAt, stock.CreatedAt) {
								lcapGain += gain
							} else {
								scapGain += gain
							}
						}
						if qty == 0 {
							break
						}
					}
				}
				if indexToPop != -1 {
					profitsMap[stockTicker] = profitsMap[stockTicker][indexToPop+1:]
				}
				if lcapGain != 0.0 {
					profit := Profit{
						Date:   strings.Split(stock.CreatedAt, " ")[0],
						Amount: lcapGain,
						Lcap:   true,
						Ticker: stockTicker,
						Tag:    stock.Tag,
					}
					profitList = append(profitList, profit)
				}
				if scapGain != 0.0 {
					profit := Profit{
						Date:   strings.Split(stock.CreatedAt, " ")[0],
						Amount: scapGain,
						Lcap:   false,
						Ticker: stockTicker,
						Tag:    stock.Tag,
					}
					profitList = append(profitList, profit)
				}
			} else { // buy
				if profitsMap[stockTicker] == nil {
					profitsMap[stockTicker] = []*models.Transaction{}
				}
				profitsMap[stockTicker] = append(profitsMap[stockTicker], &stock)
			}
		}
	}
	profitListDf := h.ConvertProfitDf(profitList)
	return profitListDf, nil
}
