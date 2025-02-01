# Picol

**Picol** is a command-line tool that provides real-time tracking of your cryptocurrency portfolio using the Binance API. It updates periodically and displays your asset balances, current prices, and 24-hour price changes in a clean and organized format.

## Features

- Real-time tracking of crypto assets via Binance WebSocket and REST API
- Periodic price and balance updates (customizable refresh interval)
- Simple and modern CLI display with color-coded price changes
- Supports **-v** or **--version** to display the current version (`v1.0.0`)
- Secure configuration using separate `env.json` (API keys) and `config.json` (settings)

## Usage

Run the `picol` command to start tracking your portfolio:

```bash
./picol
```

If you haven't configured the tool yet, you'll be prompted to provide your Binance API key and secret, as well as the refresh interval for updates. These settings will be saved in the following files:

- `~/.picol/env.json`: Contains your Binance API and secret keys.
- `~/.picol/config.json`: Contains application settings (e.g., update interval).

## Requirements

- Go 1.18+
- Binance API access (API key and secret key)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
