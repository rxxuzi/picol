// Package config internal/config/config.go
package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

// AppConfig defines the structure of config.json
type AppConfig struct {
	UpdateTime int `json:"update_time"` // リフレッシュ時間(秒)を整数として扱う
}

// GetConfigPath returns the path to config.json
func GetConfigPath() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("ユーザー情報の取得に失敗しました:", err)
		os.Exit(1)
	}
	return filepath.Join(usr.HomeDir, ".picol", "config.json")
}

// LoadAppConfig loads the app configuration from config.json
func LoadAppConfig(configPath string) (*AppConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg AppConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.UpdateTime <= 0 {
		cfg.UpdateTime = 5
	}
	return &cfg, nil
}

// CreateAppConfigInteractive creates config.json by prompting the user for input
func CreateAppConfigInteractive(configPath string) (*AppConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("config.jsonが見つかりません。アプリ設定を開始します。")
	fmt.Print("Update time (秒) (未入力なら 5): ")
	updateTimeStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	updateTimeStr = strings.TrimSpace(updateTimeStr)

	updateTime := 5 // default
	if updateTimeStr != "" {
		if val, err := strconv.Atoi(updateTimeStr); err == nil && val > 0 {
			updateTime = val
		} else {
			fmt.Println("数値以外が入力されたか、0以下でした。デフォルト 5 秒を適用します。")
		}
	}

	cfg := AppConfig{
		UpdateTime: updateTime,
	}

	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, err
		}
	}

	file, err := os.Create(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&cfg); err != nil {
		return nil, err
	}

	fmt.Println("config.jsonが作成されました:", configPath)
	return &cfg, nil
}
