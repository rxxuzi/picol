package display

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/rxxuzi/picol/internal/client"
)

var previousPrices = make(map[string]float64)

type Portfolio struct {
	Balances     []binance.Balance
	PriceInfo    map[string]client.PriceInfo
	TotalBalance float64
	Mu           sync.Mutex
}

func DisplayPortfolioModern(portfolio *Portfolio, refreshSeconds int) {
	portfolio.Mu.Lock()
	defer portfolio.Mu.Unlock()

	// clear
	fmt.Print("\033[H\033[2J")

	// header
	fmt.Printf("Balance (USD): $%.2f, Time %ds\n\n", portfolio.TotalBalance, refreshSeconds)
	headerFmt := "%-6s | %12s | %14s | %11s | %16s\n"
	fmt.Printf(headerFmt, "COIN", "AMOUNT", "PRICE (USD)", "USD VALUE", "24H CHANGES (%)")
	fmt.Println("---------------------------------------------------------------")

	rowFmt := "%-6s | %12s | %14s | %11s | %16s\n"

	for _, balance := range portfolio.Balances {
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
		priceInfo, ok := portfolio.PriceInfo[symbol]
		if !ok {
			continue
		}

		priceSymbol := compareAndGetSymbol(symbol, priceInfo.Price)
		priceStr := fmt.Sprintf("%s%.2f", priceSymbol, priceInfo.Price)

		usdtValue := total * priceInfo.Price
		changeStr := getChangeText(priceInfo.Change24h)

		fmt.Printf(rowFmt,
			asset,
			fmt.Sprintf("%.8f", total),
			priceStr,
			fmt.Sprintf("%.2f", usdtValue),
			changeStr,
		)
	}
}

func compareAndGetSymbol(symbol string, newPrice float64) string {
	oldPrice, exists := previousPrices[symbol]
	if !exists {
		previousPrices[symbol] = newPrice
		return "▶"
	}
	if newPrice > oldPrice {
		previousPrices[symbol] = newPrice
		return "▲"
	} else if newPrice < oldPrice {
		previousPrices[symbol] = newPrice
		return "▼"
	}
	return "▶"
}

func getChangeText(change float64) string {
	changeStr := fmt.Sprintf("%.2f%%", change)
	if change > 0 {
		return text.FgGreen.Sprint(changeStr)
	} else if change < 0 {
		return text.FgRed.Sprint(changeStr)
	}
	return changeStr
}
