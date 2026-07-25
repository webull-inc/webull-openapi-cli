# Webull OpenAPI CLI

CLI for [Webull](https://www.webull.com/) OpenAPI. Trade stocks, options, futures, crypto, and event contracts, access market data, and manage your account from the terminal.

> **Alpha Preview** — This CLI is under active development. Commands, flags, and output formats may change between releases.

## Features

- **Full asset coverage**: Stocks, options, futures, crypto, event contracts
- **Multi-region**: US, HK, JP, SG, TH, AU, MY, UK, BR, MX, EU, ZA
- **Agent-first design**: JSON output, `--quiet` mode, structured errors, explicit exit codes
- **AK/SK authentication**: HMAC signature + optional 2FA token
- **Safety**: Order confirmation prompts (skip with `-y`), `--dry-run` mode
- **Scriptable**: Pipe-friendly, `--csv` export

## Install

**Download pre-built binary (recommended):**

Download the latest binary for your platform from [Releases](https://github.com/webull-inc/webull-openapi-cli/releases), extract, and add to your PATH.

**Go install (requires Go 1.26.2+):**

```bash
go install github.com/webull-inc/webull-openapi-cli/cmd/webull@latest
```

**From source:**

```bash
git clone https://github.com/webull-inc/webull-openapi-cli.git
cd webull-openapi-cli
go build ./cmd/webull/
# Produces: webull (macOS/Linux) or webull.exe (Windows)
```

**Add to PATH (so you can use `webull` directly):**

```bash
# macOS/Linux: add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$HOME/go/bin"    # if using go install
export PATH="$PATH:/path/to/bin"    # if using make build

# Windows (PowerShell):
$env:PATH += ";C:\path\to\webull"
# Or move webull.exe to a directory already in PATH
```

## Quick Start

```bash
# Authenticate
webull auth login --region us

# Check your account
webull account list

# Get a stock quote
webull data stock snapshot --symbol AAPL

# Place an order
webull order stock submit --account-id ACC123 --symbol AAPL --side buy --qty 10 --type market -y

# View positions
webull account positions --account-id ACC123

# Query instrument info
webull instrument stock --symbol AAPL

# Check top gainers
webull data screener gainers-losers --direction DESC

# Get company fundamentals
webull data fundamentals company-profile --symbol AAPL
webull data fundamentals earnings-calendar --symbol 00700 --category HK_STOCK
```

## Authentication

### AK/SK + 2FA (Phase 1)

```bash
webull auth login --region us
```

Prompts for App Key and App Secret. If 2FA is enabled, you'll be asked to verify in the Webull App.

Credentials (AK/SK and 2FA token) are stored securely in the system keychain:
- **macOS**: Keychain Access (service: `webull-cli`)
- **Windows**: Credential Manager
- **Linux**: Secret Service (gnome-keyring / KDE Wallet)

Non-sensitive profile data (region, env, endpoints) is stored in `~/.config/webull/profiles.json` (macOS/Linux) or `%USERPROFILE%\.config\webull\profiles.json` (Windows).

### UAT Environment

For non-production environments, you must provide endpoint hosts explicitly:

```bash
webull auth login --region hk --env uat \
  --api-endpoint api.sandbox.webull.hk \
  --name hk-sandbox
```

`--api-endpoint` is required for UAT.

### Environment Variables

For CI/automation, use environment variables instead of interactive login:

```bash
export WEBULL_APP_KEY=your-app-key
export WEBULL_APP_SECRET=your-app-secret
webull account list
```

### Credential Precedence

1. Environment variables (`WEBULL_APP_KEY` + `WEBULL_APP_SECRET`)
2. `--profile` flag
3. Active profile from config file

## Commands

### Trading (`webull order`)

Orders are organized by asset class under `webull order <asset>`:

| Command | Description |
|---------|-------------|
| `webull order stock submit` | Place a stock/ETF order |
| `webull order stock preview` | Preview stock order fees |
| `webull order stock replace` | Modify a stock order |
| `webull order stock combo` | Place a combo order (OTO/OCO/OTOCO) via JSON |
| `webull order option submit` | Place a single-leg option order |
| `webull order option preview` | Preview option order fees |
| `webull order option replace` | Modify an option order |
| `webull order option strategy` | Place a multi-leg option strategy via JSON |
| `webull order futures submit` | Place a futures order |
| `webull order futures replace` | Modify a futures order |
| `webull order crypto submit` | Place a crypto order |
| `webull order event submit` | Place an event contract order |
| `webull order event replace` | Modify an event contract order |
| `webull order cancel` | Cancel an order (any asset) |
| `webull order open` | List open orders |
| `webull order detail` | Get order details |
| `webull order history` | Order history |

### Market Data (`webull data`)

Data is organized by asset class under `webull data <asset>`:

| Command | Description |
|---------|-------------|
| `webull data stock snapshot` | Real-time stock snapshot |
| `webull data stock bars` | Historical candlestick bars |
| `webull data stock batch-bars` | Batch historical bars (multiple symbols) |
| `webull data stock quotes` | Depth quotes (bid/ask) |
| `webull data stock tick` | Tick-by-tick trades |
| `webull data stock footprint` | Large order footprint |
| `webull data stock noii-bars` | NOII (Net Order Imbalance) bars |
| `webull data stock noii-snapshot` | NOII snapshot |
| `webull data option snapshot` | Option real-time snapshot |
| `webull data option bars` | Option historical bars |
| `webull data option tick` | Option tick data |
| `webull data futures snapshot` | Futures snapshot |
| `webull data futures bars` | Futures historical bars |
| `webull data futures tick` | Futures tick data |
| `webull data futures footprint` | Futures footprint data |
| `webull data futures depth` | Futures depth of book |
| `webull data crypto snapshot` | Crypto snapshot |
| `webull data crypto bars` | Crypto historical bars |
| `webull data event snapshot` | Event contract snapshot |
| `webull data event bars` | Event contract bars |
| `webull data event tick` | Event contract tick data |
| `webull data event depth` | Event contract order book |
| `webull data screener gainers-losers` | Top gainers/losers |
| `webull data screener actives` | Most active stocks |
| `webull data screener market-sectors` | Market sector overview and rankings |
| `webull data screener market-sectors-detail` | Stock list for a specific sector |
| `webull data screener high-dividend` | High dividend stocks ranking |
| `webull data screener 52whl` | 52-week high/low stocks |
| `webull data fundamentals company-profile` | Company profile information |
| `webull data fundamentals analyst-price` | Analyst target price |
| `webull data fundamentals analyst-rating` | Analyst rating |
| `webull data fundamentals forecast-eps` | Forecast EPS (most recent 5 quarters) |
| `webull data fundamentals filings` | SEC filings |
| `webull data fundamentals earnings-calendar` | Earnings calendar |
| `webull data fundamentals dividend-calendar` | Dividend calendar |
| `webull data fundamentals capital-flow` | Capital flow distribution |
| `webull data fundamentals industry-comparison` | Industry peer comparison |
| `webull data fundamentals indicators` | Financial indicators (ROA, ROE, EPS) |
| `webull data fundamentals income` | Income statement |
| `webull data fundamentals cashflow` | Cash flow statement |
| `webull data fundamentals balancesheet` | Balance sheet |
| `webull data fundamentals alert` | Earnings release alert |
| `webull data fundamentals fund-brief` | Fund brief information |
| `webull data fundamentals fund-performance` | Fund performance returns |
| `webull data fundamentals fund-net-value` | Fund NAV history |
| `webull data fundamentals fund-holdings` | Fund top 10 holdings |
| `webull data fundamentals fund-dividends` | Fund dividend history |
| `webull data fundamentals fund-rating` | Fund rating history |
| `webull data fundamentals fund-splits` | Fund split history |
| `webull data fundamentals fund-files` | Fund documents and files |
| `webull data fundamentals fund-allocation` | Fund asset allocation |

### Instruments (`webull instrument`)

| Command | Description |
|---------|-------------|
| `webull instrument stock` | Query stock/ETF instruments |
| `webull instrument crypto` | Query crypto instruments |
| `webull instrument futures` | Query futures contracts |
| `webull instrument futures-products` | Futures product codes |
| `webull instrument futures-class` | Futures product classes |
| `webull instrument event` | Query event contract instruments |
| `webull instrument event-categories` | List event categories |
| `webull instrument event-series` | List event series |
| `webull instrument event-events` | List events by series |
| `webull instrument company-profile` | Company profile *(deprecated, use `data fundamentals`)* |
| `webull instrument analyst-price` | Analyst target price *(deprecated, use `data fundamentals`)* |
| `webull instrument analyst-rating` | Analyst rating *(deprecated, use `data fundamentals`)* |

### Watchlists (`webull watchlist`)

| Command | Description |
|---------|-------------|
| `webull watchlist list` | List all watchlists |
| `webull watchlist get` | Get watchlist instruments |
| `webull watchlist create` | Create a watchlist |
| `webull watchlist update` | Update watchlist name/sort |
| `webull watchlist delete` | Delete a watchlist |
| `webull watchlist add` | Add instruments (JSON) |
| `webull watchlist remove` | Remove instruments |
| `webull watchlist sort` | Update instruments sort order |

### Account

| Command | Description |
|---------|-------------|
| `webull account list` | List trading accounts |
| `webull account balance` | Account balance |
| `webull account positions` | Account positions |
| `webull account position-detail` | Position details by contract (JP) |

### Profile Management

| Command | Description |
|---------|-------------|
| `webull auth login` | Authenticate (AK/SK + 2FA) |
| `webull auth status` | Check auth status |
| `webull profile list` | List profiles |
| `webull profile switch` | Switch active profile |
| `webull profile logout` | Remove a profile |
| `webull profile set-endpoint` | Override API endpoint |

### Utilities

| Command | Description |
|---------|-------------|
| `webull doctor` | Diagnose config and connectivity |
| `webull version` | Print version |

Every command supports `--help` for full flag documentation:

```bash
webull order stock submit --help    # See all flags for stock submit
webull data stock bars --help       # See all flags for bars query
```

### JSON Input Mode

For complex orders (combo, strategy, or orders with array fields like `close_contracts`), use `--body` or `--file`:

```bash
# Inline JSON
webull order stock combo --body '{"account_id":"ACC123","new_orders":[...]}'

# From file
webull order option strategy --file my_strategy.json

# Pipe from stdin
cat order.json | webull order stock combo
```

## Output Formats

```bash
webull account positions --account-id ACC123            # JSON (default)
webull account positions --account-id ACC123 --csv     # CSV
webull data stock snapshot --symbol AAPL,TSLA          # Multiple symbols
```

## Safety

- **Order confirmation**: `order stock submit`, `order cancel`, and `order stock replace` prompt for confirmation in interactive mode. Use `-y` to skip.
- **Dry run**: `webull order stock submit --dry-run` prints the request payload without submitting.
- **Credentials**: AK/SK and tokens stored in system keychain (not in plaintext files). Non-sensitive config uses `0600` file permissions.

## Global Flags

| Flag | Description |
|------|-------------|
| `--profile` | Use a specific profile |
| `--region` | Override profile region |
| `--env` | Environment (prod/uat) |
| `--csv` | Output in CSV format |
| `--quiet` | Suppress non-data output |
| `--verbose` | Show request summaries on stderr |
| `--debug` | Show full request/response on stderr |
| `--timeout` | HTTP timeout in seconds (default: 30) |
| `-y, --yes` | Skip confirmation prompts |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Any error |

Errors are output as JSON to stderr with an `error_code` field for machine parsing:

```json
{"error_code":"CLI_ERROR","message":"required flag(s) \"qty\", \"side\", \"type\" not set"}
```

## Development

```bash
make build          # Build binary to bin/webull
make test           # Run tests
make lint           # Run linter
make clean          # Remove build artifacts
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

## Disclaimer

See [DISCLAIMER.md](DISCLAIMER.md) for important risk warnings and terms of use.
