// Package client internal/client/client.go
package client

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"strconv"
	"sync"
)

// PriceInfo holds the price and 24h change percentage
type PriceInfo struct {
	Price     float64
	Change24h float64
}

// BinanceClient wraps the go-binance client and additional data
type BinanceClient struct {
	Client     *binance.Client
	PriceInfo  map[string]PriceInfo
	PriceMutex sync.Mutex
}

// NewBinanceClient initializes and returns a BinanceClient
func NewBinanceClient(apiKey, secretKey string) *BinanceClient {
	client := binance.NewClient(apiKey, secretKey)
	return &BinanceClient{
		Client:    client,
		PriceInfo: make(map[string]PriceInfo),
	}
}

// FetchAccountInfo retrieves account balances
func (bc *BinanceClient) FetchAccountInfo() ([]binance.Balance, error) {
	account, err := bc.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	var balances []binance.Balance
	for _, balance := range account.Balances {
		free, err1 := strconv.ParseFloat(balance.Free, 64)
		locked, err2 := strconv.ParseFloat(balance.Locked, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if free != 0 || locked != 0 {
			balances = append(balances, balance)
		}
	}
	return balances, nil
}

// FetchPriceAndChange retrieves price and 24h change for given symbols
func (bc *BinanceClient) FetchPriceAndChange(symbols []string) (map[string]PriceInfo, error) {
	prices, err := bc.Client.NewListPricesService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	priceMap := make(map[string]float64)
	for _, p := range prices {
		priceMap[p.Symbol] = parseFloatOrZero(p.Price)
	}

	priceChangeStats, err := bc.Client.NewListPriceChangeStatsService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	changeMap := make(map[string]float64)
	for _, c := range priceChangeStats {
		changeMap[c.Symbol] = parseFloatOrZero(c.PriceChangePercent)
	}

	result := make(map[string]PriceInfo)
	for _, symbol := range symbols {
		price, ok1 := priceMap[symbol]
		change, ok2 := changeMap[symbol]
		if ok1 && ok2 {
			result[symbol] = PriceInfo{
				Price:     price,
				Change24h: change,
			}
		}
	}

	// スレッドセーフに更新
	bc.PriceMutex.Lock()
	defer bc.PriceMutex.Unlock()
	bc.PriceInfo = result

	return result, nil
}

// parseFloatOrZero parses a string to float64, returns 0 on failure
func parseFloatOrZero(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
