package rhwrapper

// package for interacting with robinhood API

import (
	"container/list"
	"context"
	"fmt"
	"os"

	robinhood "github.com/Ryang20718/robinhood-client/client"
	models "github.com/Ryang20718/robinhood-client/models"
)

type HoodAPI interface {
	Auth(username string, password string, mfa string) error
	FetchOptionTrades(ctx context.Context) (map[string][]models.OptionTransaction, error)
	FetchRegularTrades(ctx context.Context) (map[string][]models.Transaction, error)
}

type Hood struct {
	cli *robinhood.Client
}

func (h *Hood) Auth(username string, password string, mfa string) error {
	if username == "" {
		return fmt.Errorf("requires a username")
	}
	if password == "" {
		return fmt.Errorf("requires a password")
	}
	if mfa == "" {
		return fmt.Errorf("requires an mfa")
	}
	cli, err := robinhood.Dial(
		&robinhood.OAuth{
			Username: os.Getenv(username),
			Password: os.Getenv(password),
			MFA:      mfa,
		})
	if err != nil {
		return fmt.Errorf("failed to auth rhood err: %v", err.Error())
	}
	h.cli = cli
	return nil
}

/*
Returns mapping of ticker to model.OptionTransaction

Golang stdlib list is kinda dumb
*/
func (h *Hood) FetchOptionTrades(ctx context.Context) (map[string]*list.List, error) {
	optionsOrderMap := make(map[string]*list.List)
	optionOrders, err := h.cli.GetOptionsOrders(ctx)
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
		optionsOrderMap[order.Ticker].PushBack(order)
	}
	return optionsOrderMap, nil
}

/*
Returns mapping of ticker to model.Transaction

Golang stdlib list is kinda dumb
*/
func (h *Hood) FetchRegularTrades(ctx context.Context) (map[string]*list.List, error) {
	stockOrderMap := make(map[string]*list.List)
	stockOrders, err := h.cli.GetStockOrders()
	if err != nil {
		return nil, err
	}
	for _, order := range stockOrders {
		if stockOrderMap[order.Ticker] == nil {
			stockOrderMap[order.Ticker] = list.New()
		}
		stockOrderMap[order.Ticker].PushBack(order)
	}
	return stockOrderMap, nil
}

// func (h *Hood) ProcessRealizedEarnings(ctx context.Context) error {
// 	stockOrderMap, err := h.FetchRegularTrades(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	optionOrderMap, err := h.FetchOptionTrades(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
