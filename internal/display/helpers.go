// Package display internal/display/helpers.go
package display

import (
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/rxxuzi/picol/internal/client"
)

func MergeBalances(existing, updated []binance.Balance) []binance.Balance {
	balanceMap := make(map[string]binance.Balance)
	for _, balance := range existing {
		balanceMap[balance.Asset] = balance
	}
	for _, balance := range updated {
		balanceMap[balance.Asset] = balance
	}
	var merged []binance.Balance
	for _, balance := range balanceMap {
		merged = append(merged, balance)
	}
	return merged
}

func CalculateTotalBalance(balances []binance.Balance, priceInfo map[string]client.PriceInfo) float64 {
	var totalBalance float64 = 0
	for _, balance := range balances {
		asset := balance.Asset
		free, err1 := strconv.ParseFloat(balance.Free, 64)
		locked, err2 := strconv.ParseFloat(balance.Locked, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		total := free + locked
		if total == 0 {
			continue
		}

		symbol := asset + "USDT"
		price, ok := priceInfo[symbol]
		if !ok {
			continue
		}
		totalBalance += total * price.Price
	}
	return totalBalance
}
