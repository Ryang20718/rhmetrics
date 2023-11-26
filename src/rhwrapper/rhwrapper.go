package rhwrapper

// package for interacting with robinhood API

import (
	"container/list"
	"context"
	"fmt"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"

	robinhood "github.com/Ryang20718/robinhood-client/client"
	models "github.com/Ryang20718/robinhood-client/models"
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

# Golang stdlib list is kinda dumb

Each ticker maps to a list which is sorted by created datetime
*/
func (h *Hood) FetchOptionTrades(ctx context.Context) (map[string]*list.List, error) {
	optionsOrderMap := make(map[string]*list.List)
	optionOrders, err := h.Cli.GetOptionsOrders(ctx)
	if err != nil {
		return nil, err
	}
	if optionOrders == nil {
		return optionsOrderMap, nil
	}
	for _, order := range *optionOrders {
		if optionsOrderMap[order.Ticker] == nil {
			optionsOrderMap[order.Ticker] = list.New()
		}
		optionsOrderMap[order.Ticker].PushFront(order)
	}
	return optionsOrderMap, nil
}

/*
Returns mapping of ticker to model.Transaction

# Golang stdlib list is kinda dumb

Each ticker maps to a list which is sorted by created datetime
*/
func (h *Hood) FetchRegularTrades(ctx context.Context) (map[string]*list.List, error) {
	stockOrderMap := make(map[string]*list.List)
	stockOrders, err := h.Cli.GetStockOrders()
	if err != nil {
		return nil, err
	}
	for _, order := range stockOrders {
		if stockOrderMap[order.Ticker] == nil {
			stockOrderMap[order.Ticker] = list.New()
		}
		stockOrderMap[order.Ticker].PushFront(order)
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

/*
convert transaction to dataframe
*/
func (h *Hood) ConvertTransactionDF(ctx context.Context) error {
	stockOrderMap, err := h.FetchRegularTrades(ctx)
	if err != nil {
		return err
	}

	tickers := series.String([]string{})
	transactionType := series.String([]string{})
	qty := series.float64([]float64{})
	unitCost := series.Float([]float64{})
	createdAt := series.Time([]time.Time{})
	tag := series.String([]string{})

	// Populate series with data from struct array
	for _, stockOrders := range stockOrderMap {
		for order := stockOrders.Front(); order != nil; order = order.Next() {

			tickers.Append(order.Ticker)
			transactionType.Append(order.TransactionType)
			qty.Append(order.Qty)
			unitCost.Append(order.UnitCost)
			createdAt.Append(order.CreatedAt)
			tag.Append(order.Tag)
		}
	}
	df := dataframe.New(
		tickers,
		transactionType,
		qty,
		unitCost,
		createdAt,
		tag,
	)
}

// /*
// convert transaction to dataframe
// */
// func (h *Hood) ConvertOptionTransactionDF(ctx context.Context) error {

// }

func (h *Hood) ProcessRealizedEarnings(ctx context.Context) error {

	// optionOrderMap, err := h.FetchOptionTrades(ctx)
	// if err != nil {
	// 	return err
	// }

	return nil
}
