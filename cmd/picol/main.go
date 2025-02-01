package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/rxxuzi/picol/internal/client"
	"github.com/rxxuzi/picol/internal/config"
	"github.com/rxxuzi/picol/internal/display"
)

const Version = "v1.0.0"

func main() {
	if checkVersionFlag(os.Args) {
		fmt.Printf("picol %s\n", Version)
		return
	}

	envPath := config.GetEnvPath()
	var envConfig *config.EnvConfig
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		ec, err := config.CreateEnvConfigInteractive(envPath)
		if err != nil {
			log.Fatalf("env.jsonの作成に失敗: %v", err)
		}
		envConfig = ec
	} else {
		ec, err := config.LoadEnvConfig(envPath)
		if err != nil {
			log.Fatalf("env.jsonの読み込みに失敗: %v", err)
		}
		envConfig = ec
	}

	configPath := config.GetConfigPath()
	var appConfig *config.AppConfig
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		ac, err := config.CreateAppConfigInteractive(configPath)
		if err != nil {
			log.Fatalf("config.jsonの作成に失敗: %v", err)
		}
		appConfig = ac
	} else {
		ac, err := config.LoadAppConfig(configPath)
		if err != nil {
			log.Fatalf("config.jsonの読み込みに失敗: %v", err)
		}
		appConfig = ac
	}

	// リフレッシュ秒数を取得
	refreshSeconds := appConfig.UpdateTime
	if refreshSeconds <= 0 {
		refreshSeconds = 5
	}

	// Binanceクライアント初期化
	binanceClient := client.NewBinanceClient(envConfig.APIKey, envConfig.SecretKey)

	// 初回アカウント情報
	balances, err := binanceClient.FetchAccountInfo()
	if err != nil {
		log.Fatalf("アカウント情報の取得失敗: %v", err)
	}

	var symbols []string
	for _, b := range balances {
		symbols = append(symbols, b.Asset+"USDT")
	}

	// 初回価格情報
	priceInfo, err := binanceClient.FetchPriceAndChange(symbols)
	if err != nil {
		log.Fatalf("価格情報の取得失敗: %v", err)
	}

	// ポートフォリオ作成
	portfolio := &display.Portfolio{
		Balances:     balances,
		PriceInfo:    priceInfo,
		TotalBalance: display.CalculateTotalBalance(balances, binanceClient.PriceInfo),
	}

	// 初回表示
	display.DisplayPortfolioModern(portfolio, refreshSeconds)

	// リアルタイム更新(WebSocket)用チャネル
	updateChan := make(chan []binance.Balance)

	// シグナル(Ctrl+C等)を受け取ったら終了
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go startWebSocket(binanceClient.Client, updateChan)

	ticker := time.NewTicker(time.Duration(refreshSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case updatedBalances := <-updateChan:
			portfolio.Mu.Lock()
			portfolio.Balances = display.MergeBalances(portfolio.Balances, updatedBalances)
			var newSymbols []string
			for _, b := range portfolio.Balances {
				newSymbols = append(newSymbols, b.Asset+"USDT")
			}
			newPrice, err := binanceClient.FetchPriceAndChange(newSymbols)
			if err != nil {
				log.Printf("価格情報の再取得失敗: %v", err)
				portfolio.Mu.Unlock()
				continue
			}
			portfolio.PriceInfo = newPrice
			portfolio.TotalBalance = display.CalculateTotalBalance(portfolio.Balances, binanceClient.PriceInfo)
			portfolio.Mu.Unlock()

			display.DisplayPortfolioModern(portfolio, refreshSeconds)

		case <-ticker.C:
			portfolio.Mu.Lock()
			var currentSymbols []string
			for _, b := range portfolio.Balances {
				currentSymbols = append(currentSymbols, b.Asset+"USDT")
			}
			newPrice, err := binanceClient.FetchPriceAndChange(currentSymbols)
			if err != nil {
				log.Printf("価格情報の定期更新失敗: %v", err)
				portfolio.Mu.Unlock()
				continue
			}
			portfolio.PriceInfo = newPrice
			portfolio.TotalBalance = display.CalculateTotalBalance(portfolio.Balances, binanceClient.PriceInfo)
			portfolio.Mu.Unlock()

			display.DisplayPortfolioModern(portfolio, refreshSeconds)

		case sig := <-sigChan:
			fmt.Printf("受信シグナル: %v. アプリケーションを終了します。\n", sig)
			return
		}
	}
}

// startWebSocket listens for account updates and sends updated balances to the channel.
func startWebSocket(client *binance.Client, updateChan chan<- []binance.Balance) {
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Listen key取得失敗: %v", err)
	}

	wsHandler := func(event *binance.WsUserDataEvent) {
		if len(event.AccountUpdate.WsAccountUpdates) > 0 {
			var updatedBalances []binance.Balance
			for _, ua := range event.AccountUpdate.WsAccountUpdates {
				free, err1 := strconv.ParseFloat(ua.Free, 64)
				locked, err2 := strconv.ParseFloat(ua.Locked, 64)
				if err1 != nil || err2 != nil {
					continue
				}
				if free != 0 || locked != 0 {
					updatedBalances = append(updatedBalances, binance.Balance{
						Asset:  ua.Asset,
						Free:   ua.Free,
						Locked: ua.Locked,
					})
				}
			}
			if len(updatedBalances) > 0 {
				updateChan <- updatedBalances
			}
		}
	}

	errHandler := func(err error) {
		log.Printf("WebSocketエラー: %v", err)
	}

	doneC, stopC, err := binance.WsUserDataServe(listenKey, wsHandler, errHandler)
	if err != nil {
		log.Fatalf("WebSocket接続失敗: %v", err)
	}
	defer close(stopC)

	<-doneC
}

func checkVersionFlag(args []string) bool {
	for _, a := range args {
		if a == "-v" || a == "--version" {
			return true
		}
	}
	return false
}
