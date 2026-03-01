# etoro-cli

> The command-line Swiss Army knife for eToro trading automation

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/lijinlar/etoro-cli/pulls)

---

## About

**etoro-cli** is a production-quality command-line interface for eToro trading automation. Built with safety-first principles and designed to be both human-friendly and agent-friendly (AI agents love the `--json` flag!).

### Why etoro-cli?

- **Full Trading Capabilities**: View account, place orders, manage positions, track portfolio
- **Agent-Friendly**: JSON output mode for seamless integration with AI agents and scripts
- **Safety First**: Kill switch, execution toggles, trade limits, and symbol allowlists
- **Comprehensive**: 13 commands covering everything from price checks to risk analysis
- **Clean Output**: Beautiful tables for humans, structured JSON for machines

---

## Installation

Works on **Linux**, **macOS**, and **Windows**.

### Pre-built Binaries (Fastest)

Download the latest release for your platform from the [Releases page](https://github.com/lijinlar/etoro-cli/releases).

**Linux / macOS:**
```bash
# Linux (amd64)
curl -L https://github.com/lijinlar/etoro-cli/releases/latest/download/etoro-cli-linux-amd64.tar.gz | tar xz
sudo mv etoro-cli /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/lijinlar/etoro-cli/releases/latest/download/etoro-cli-darwin-arm64.tar.gz | tar xz
sudo mv etoro-cli /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/lijinlar/etoro-cli/releases/latest/download/etoro-cli-darwin-amd64.tar.gz | tar xz
sudo mv etoro-cli /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/lijinlar/etoro-cli/releases/latest/download/etoro-cli-windows-amd64.zip" -OutFile etoro-cli.zip
Expand-Archive etoro-cli.zip -DestinationPath .
```

### Using Go Install

```bash
go install github.com/lijinlar/etoro-cli@latest
```

### Building from Source

```bash
git clone https://github.com/lijinlar/etoro-cli.git
cd etoro-cli

# Current platform
make build

# All platforms (outputs to ./dist/)
make build-all
```

---

## Quick Start

### 1. Create Configuration

```bash
cp config.example.yaml etoro.yaml
```

### 2. Add Your API Keys

Edit `etoro.yaml`:

```yaml
etoro:
  public_key: "your-public-key"
  user_key: "your-user-key"
```

### 3. Check Your Account

```bash
etoro account
```

### 4. Ready to Trade?

First, disable the safety guards:

```yaml
trading:
  execution_enabled: true
  kill_switch: false
```

Then execute with `--confirm`:

```bash
etoro buy AAPL --amount 100 --confirm
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `account` | Show account summary | `etoro account` |
| `price` | Get live prices | `etoro price AAPL TSLA` |
| `positions` | List open positions | `etoro positions --symbol AAPL` |
| `orders` | List pending orders | `etoro orders` |
| `buy` | Place buy order | `etoro buy AAPL --amount 100 --confirm` |
| `sell` | Place sell (short) order | `etoro sell TSLA --amount 200 --confirm` |
| `close` | Close position | `etoro close 12345 --confirm` |
| `cancel` | Cancel pending order | `etoro cancel 67890 --confirm` |
| `portfolio` | Portfolio snapshot | `etoro portfolio` |
| `watchlist` | Manage watchlist | `etoro watchlist --add GOOGL` |
| `history` | Trade history | `etoro history --from 2024-01-01` |
| `risk` | Risk dashboard | `etoro risk` |
| `search` | Search instruments | `etoro search bitcoin` |

### Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Path to config file |
| `--json` | Output as JSON (for agents/scripts) |
| `--dry-run` | Simulate without executing |
| `--verbose` | Show HTTP requests/responses |

---

## Agent Usage

etoro-cli is designed for seamless integration with AI agents and automation scripts. Use the `--json` flag for structured output.

### JSON Output Examples

```bash
# Account info as JSON
etoro account --json

# Price data for parsing
etoro price AAPL TSLA --json | jq '.[] | {symbol, bid, ask}'

# Position data
etoro positions --json | jq '.[] | select(.pl > 0)'
```

### Scripting Patterns

#### Check Position P&L

```bash
#!/bin/bash
PL=$(etoro positions --json | jq '[.[] | .pl] | add')
echo "Total unrealized P&L: $PL"
```

#### Automated Risk Check

```bash
#!/bin/bash
MARGIN=$(etoro risk --json | jq '.marginUtilization')
if (( $(echo "$MARGIN > 70" | bc -l) )); then
    echo "WARNING: High margin utilization!"
fi
```

#### Place Order with Confirmation

```bash
# Dry-run first
etoro buy AAPL --amount 100 --json

# If looks good, execute
ETORO_CONFIRM=1 etoro buy AAPL --amount 100
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API error |
| 2 | Configuration error |
| 3 | Validation error |

---

## Configuration Reference

Configuration is loaded from:
1. Path specified by `--config` flag
2. `./etoro.yaml` (current directory)
3. `~/.etoro/config.yaml` (home directory)

Environment variables override config file values.

### Full Configuration

| Key | Env Variable | Description | Default |
|-----|--------------|-------------|---------|
| `etoro.public_key` | `ETORO_PUBLIC_KEY` | eToro public API key | (required) |
| `etoro.user_key` | `ETORO_USER_KEY` | eToro user API key | (required) |
| `etoro.base_url` | `ETORO_BASE_URL` | API base URL | `https://api.etoro.com` |
| `trading.execution_enabled` | `ETORO_EXECUTION_ENABLED` | Allow trade execution | `false` |
| `trading.kill_switch` | `ETORO_KILL_SWITCH` | Emergency stop all trading | `true` |
| `trading.max_trade_usd` | `ETORO_MAX_TRADE_USD` | Max USD per trade | `500` |
| `trading.max_positions` | - | Max concurrent positions | `10` |
| `trading.symbol_allowlist` | - | Allowed symbols (empty=all) | `[]` |
| `trading.default_leverage` | - | Default leverage | `1` |
| `output.format` | - | Output format (table/json) | `table` |
| `output.timezone` | - | Timestamp timezone | `local` |

---

## Safety Guardrails

etoro-cli implements multiple safety layers to prevent accidental trades:

### 1. Kill Switch

The nuclear option. When enabled, **ALL** trading operations fail immediately.

```yaml
trading:
  kill_switch: true  # Default: ON
```

Error: `Kill switch is active. Set kill_switch: false in config to enable trading.`

### 2. Execution Toggle

Global toggle for actual trade execution.

```yaml
trading:
  execution_enabled: false  # Default: OFF
```

When disabled, trading commands show dry-run previews only.

### 3. Maximum Trade Amount

Prevents accidentally large orders.

```yaml
trading:
  max_trade_usd: 500  # Reject orders > $500
```

Error: `order amount $1000.00 exceeds max_trade_usd limit of $500.00`

### 4. Symbol Allowlist

Restrict trading to specific instruments.

```yaml
trading:
  symbol_allowlist: ["AAPL", "GOOGL", "MSFT"]
```

Error: `Symbol TSLA is not in the allowed list.`

### 5. Confirmation Requirement

Trading commands require explicit confirmation:

```bash
# Dry-run (no --confirm)
etoro buy AAPL --amount 100

# Execute (with --confirm)
etoro buy AAPL --amount 100 --confirm

# Or via environment variable
ETORO_CONFIRM=1 etoro buy AAPL --amount 100
```

---

## Examples

### View Account Status

```bash
$ etoro account

+-------------------+------------+
| FIELD             | VALUE      |
+-------------------+------------+
| Login ID          | user123    |
| Balance           | $10,000.00 |
| Equity            | $12,500.00 |
| Margin            | $2,500.00  |
| Available Margin  | $10,000.00 |
| Unrealized P&L    | +$2,500.00 |
| Realized P&L Today| +$150.00   |
+-------------------+------------+
```

### Check Prices

```bash
$ etoro price AAPL TSLA GOOGL

+--------+---------+---------+--------+--------------+----------+----------+
| SYMBOL | BID     | ASK     | SPREAD | DAILY CHANGE | HIGH     | LOW      |
+--------+---------+---------+--------+--------------+----------+----------+
| AAPL   | $178.50 | $178.55 | $0.05  | +1.25%       | $180.00  | $175.00  |
| TSLA   | $245.00 | $245.10 | $0.10  | -0.50%       | $250.00  | $242.00  |
| GOOGL  | $142.30 | $142.35 | $0.05  | +0.75%       | $143.00  | $141.00  |
+--------+---------+---------+--------+--------------+----------+----------+
```

### Place a Buy Order

```bash
$ etoro buy AAPL --amount 500 --sl 170 --tp 190 --confirm

=== ORDER EXECUTED ===
Order ID:    12345
Position ID: 67890
Status:      FILLED
```

### Risk Analysis

```bash
$ etoro risk

=== RISK DASHBOARD ===
! WARNING: Margin utilization above 70% !

Margin Utilization: 72.50%
Total Exposure:     $25,000.00
Available Margin:   $3,437.50

--- Exposure by Symbol ---
  TSLA     $10,000.00 (40.0%)
  AAPL     $8,000.00 (32.0%)
  GOOGL    $7,000.00 (28.0%)
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Disclaimer

This software is provided as-is. Trading involves risk of loss. The authors are not responsible for any financial losses incurred through the use of this software. Always test with small amounts first and use the safety guardrails appropriately.

---

Built with Go and caffeineee by [LIJIN AR](https://github.com/lijinlar)
