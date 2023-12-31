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
type OptionTransaction struct {
	Ticker          string
	TransactionType string
	Qty             float64
	StrikePrice     float64
	UnitCost        float64
	CreatedAt       string
	ExpirationDate  string
	Status          string // Assigned or Expired or Open
	Tag             string
}

type Transaction struct {
	Ticker          string
	TransactionType string // Buy. Sell
	Qty             float64
	UnitCost        float64
	CreatedAt       string
	Tag             string
}
*/

// /*
// convert transaction to dataframe
// */
// func (h *Hood) ConvertTransactionDF(ctx context.Context) (*dataframe.DataFrame, error) {
// 	stockOrderMap, err := h.FetchRegularTrades(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tickers := series.New([]string{}, series.String, "tickers")
// 	transactionType := series.New([]string{}, series.String, "transactionType")
// 	qty := series.New([]float64{}, series.Float, "qty")
// 	unitCost := series.New([]float64{}, series.Float, "unitCost")
// 	createdAt := series.New([]string{}, series.String, "createdAt")
// 	tag := series.New([]string{}, series.String, "tag")

// 	// Populate series with data from struct array
// 	for _, stockOrders := range stockOrderMap {
// 		for ele := stockOrders.Front(); ele != nil; ele = ele.Next() {
// 			if order, ok := ele.Value.(models.Transaction); ok {

// 				tickers.Append(order.Ticker)
// 				transactionType.Append(order.TransactionType)
// 				qty.Append(order.Qty)
// 				unitCost.Append(order.UnitCost)
// 				createdAt.Append(order.CreatedAt)
// 				tag.Append(order.Tag)
// 			}

// 		}
// 	}
// 	df := dataframe.New(
// 		tickers,
// 		transactionType,
// 		qty,
// 		unitCost,
// 		createdAt,
// 		tag,
// 	)
// 	return &df, nil
// }

// /*
// convert transaction to dataframe
// */
// func (h *Hood) ConvertOptionTransactionDF(ctx context.Context) (*dataframe.DataFrame, error) {
// 	// Fetch option transactions
// 	optionOrderMap, err := h.FetchOptionTrades(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create series for each field
// 	tickers := series.New([]string{}, series.String, "tickers")
// 	transactionType := series.New([]string{}, series.String, "transactionType")
// 	qty := series.New([]float64{}, series.Float, "qty")
// 	strikePrice := series.New([]float64{}, series.Float, "strikePrice")
// 	unitCost := series.New([]float64{}, series.Float, "unitCost")
// 	createdAt := series.New([]string{}, series.String, "createdAt")
// 	expirationDate := series.New([]string{}, series.String, "expirationDate")
// 	status := series.New([]string{}, series.String, "status")
// 	tag := series.New([]string{}, series.String, "tag")

// 	// Populate series with data from struct array
// 	for _, optionOrders := range optionOrderMap {
// 		for ele := optionOrders.Front(); ele != nil; ele = ele.Next() {
// 			if optionTransaction, ok := ele.Value.(models.OptionTransaction); ok {

// 				tickers.Append(optionTransaction.Ticker)
// 				transactionType.Append(optionTransaction.TransactionType)
// 				qty.Append(optionTransaction.Qty)
// 				strikePrice.Append(optionTransaction.StrikePrice)
// 				unitCost.Append(optionTransaction.UnitCost)
// 				createdAt.Append(optionTransaction.CreatedAt)
// 				expirationDate.Append(optionTransaction.ExpirationDate)
// 				status.Append(optionTransaction.Status)
// 				tag.Append(optionTransaction.Tag)
// 			}
// 		}
// 	}

// 	// Create DataFrame
// 	df := dataframe.New(
// 		tickers,
// 		transactionType,
// 		qty,
// 		strikePrice,
// 		unitCost,
// 		createdAt,
// 		expirationDate,
// 		status,
// 		tag,
// 	)

// 	return &df, nil
// }

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
					Ticker:          option.Ticker,
					TransactionType: "buy",
					Qty:             100.00 * option.Qty,
					UnitCost:        option.StrikePrice + premium,
					CreatedAt:       option.ExpirationDate,
					Tag:             option.TransactionType + " assigned",
				}
				if profitsMap[option.Ticker] == nil {
					profitsMap[option.Ticker] = []*models.Transaction{}
				}
				profitsMap[option.Ticker] = append(profitsMap[option.Ticker], &stock)
			} else if option.Status == "Expired" {
				if option.TransactionType == "STO" || option.TransactionType == "STC" {
					profit := Profit{
						Date:   strings.Split(option.CreatedAt, "T")[0],
						Amount: option.Qty * 100 * option.UnitCost,
						Lcap:   false, // TODO calculate whether actual LTG
						Ticker: option.Ticker,
						Tag:    option.Tag,
					}
					profitList = append(profitList, profit)
				} else {
					profit := Profit{
						Date:   strings.Split(option.CreatedAt, "T")[0],
						Amount: -option.Qty * 100 * option.UnitCost,
						Lcap:   false, // TODO calculate whether actual LTG
						Ticker: option.Ticker,
						Tag:    option.Tag,
					}
					profitList = append(profitList, profit)
				}
			}

		} else {
			stock := stockList[stockIdx]
			stockIdx += 1
			if stock.TransactionType == "sell" {
				qty := stock.Qty
				indexToPop := -1
				lcapGain := 0.0
				scapGain := 0.0
				for qty != 0.0 {
					if len(profitsMap[stock.Ticker]) <= 0 {
						profit := Profit{
							Date:   strings.Split(stock.CreatedAt, " ")[0],
							Amount: stock.UnitCost * stock.Qty,
							Lcap:   false,
							Ticker: stock.Ticker,
							Tag:    stock.Tag,
						}
						profitList = append(profitList, profit)
						break
					}
					for i, boughtStock := range profitsMap[stock.Ticker] {
						if profitsMap[stock.Ticker][i].Qty > qty {
							gain := qty * (stock.UnitCost - boughtStock.UnitCost)
							if OneYearApart(boughtStock.CreatedAt, stock.CreatedAt) {
								lcapGain += gain
							} else {
								scapGain += gain
							}
							profitsMap[stock.Ticker][i].Qty -= qty
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
					profitsMap[stock.Ticker] = profitsMap[stock.Ticker][indexToPop+1:]
				}
				if lcapGain != 0.0 {
					profit := Profit{
						Date:   strings.Split(stock.CreatedAt, " ")[0],
						Amount: lcapGain,
						Lcap:   true,
						Ticker: stock.Ticker,
						Tag:    stock.Tag,
					}
					profitList = append(profitList, profit)
				}
				if scapGain != 0.0 {
					profit := Profit{
						Date:   strings.Split(stock.CreatedAt, " ")[0],
						Amount: scapGain,
						Lcap:   false,
						Ticker: stock.Ticker,
						Tag:    stock.Tag,
					}
					profitList = append(profitList, profit)
				}
			} else { // buy
				if profitsMap[stock.Ticker] == nil {
					profitsMap[stock.Ticker] = []*models.Transaction{}
				}
				profitsMap[stock.Ticker] = append(profitsMap[stock.Ticker], &stock)
			}
		}
	}
	profitListDf := h.ConvertProfitDf(profitList)
	return profitListDf, nil
}
