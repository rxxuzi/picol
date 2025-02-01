// Package config internal/config/env.go
package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type EnvConfig struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

func GetEnvPath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("ユーザー情報の取得に失敗しました:", err)
		os.Exit(1)
	}
	return filepath.Join(usr.HomeDir, ".picol", "env.json")
}

func LoadEnvConfig(envPath string) (*EnvConfig, error) {
	file, err := os.Open(envPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var env EnvConfig
	if err := json.NewDecoder(file).Decode(&env); err != nil {
		return nil, err
	}
	return &env, nil
}

func CreateEnvConfigInteractive(envPath string) (*EnvConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("env.jsonが見つかりません。APIキー等の設定を開始します。")

	fmt.Print("Binance API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("Binance Secret Key: ")
	secretKey, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	secretKey = strings.TrimSpace(secretKey)

	envConfig := EnvConfig{
		APIKey:    apiKey,
		SecretKey: secretKey,
	}

	configDir := filepath.Dir(envPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, err
		}
	}

	file, err := os.Create(envPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&envConfig); err != nil {
		return nil, err
	}

	fmt.Println("env.jsonが作成されました:", envPath)
	return &envConfig, nil
}
