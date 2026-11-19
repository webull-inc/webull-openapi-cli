# Webull OpenAPI CLI

CLI for [Webull](https://www.webull.com/) OpenAPI. Trade stocks, options, futures, crypto, and event contracts, access market data, and manage your account from the terminal.

> **Alpha Preview** — This CLI is under active development. Commands, flags, and output formats may change between releases.

## Features

- **Full asset coverage**: Stocks, options, futures, crypto, event contracts
- **Multi-region**: US, HK, JP, SG, TH, AU, MY, UK, BR, MX, ZA
- **Agent-first design**: JSON output, `--quiet` mode, structured errors, explicit exit codes
- **Two auth modes**: OAuth device authorization (default, browser-based) and AK/SK (HMAC signature + optional 2FA token)
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
# Authenticate (OAuth device authorization, opens a browser)
webull auth login

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

The CLI supports two authentication modes:

| Mode | Command | Credentials |
|------|---------|-------------|
| **OAuth** (default) | `webull auth login` | Browser-based device authorization |
| **AK/SK** | `webull auth login --method credentials --region <region>` | App Key / App Secret + optional 2FA |

### OAuth (default)

```bash
webull auth login
```

Runs an OAuth 2.0 device authorization flow ([RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628)):

1. The CLI prints an authorization URL and a short user code, and tries to open your browser.
2. Sign in and approve the request in the browser (choose the account and scope to authorize).
3. The CLI polls in the background and completes automatically once you approve.

```
Open the following URL in your browser to authorize:
https://.../oauth2/ai-cli/login?client_id=...&user_code=WDJB-MJHT

Enter the code: WDJB-MJHT

Browser opened. Waiting for authorization...

Successfully authenticated.
```

Notes:
- OAuth is **currently available for the US region only**. No `--region` flag is needed — the authorized region is determined during authorization.
- Works on headless machines: if no browser is available, copy the printed URL to any device to authorize.
- The access token is refreshed automatically using the refresh token; you generally only authorize once.
- PKCE is used to protect the exchange; the CLI never stores the device code.

### AK/SK + 2FA

```bash
webull auth login --method credentials --region us
```

Prompts for App Key and App Secret. If 2FA is enabled, you'll be asked to verify in the Webull App. AK/SK mode supports all regions and requires `--region`.

### Credential Storage

The CLI stores a **single** authentication record. Each successful `auth login`
overwrites the previous one, whether it was OAuth or AK/SK — there is no
multi-profile switching.

Tokens and secrets are stored securely in the system keychain:
- **macOS**: Keychain Access (service: `webull-cli`, OAuth: `webull-cli-oauth`)
- **Windows**: Credential Manager
- **Linux**: Secret Service (gnome-keyring / KDE Wallet)

**Encrypted-file fallback.** When the system keychain is unavailable (for
example a headless Linux host without Secret Service, or a CI container), the
CLI automatically falls back to an AES-256-GCM encrypted file at
`~/.config/webull/credentials-<service>.enc` (`0600`) and prints a one-line
notice to stderr. The encryption key is derived from three independent parts, so
no single artifact can decrypt the file on its own:
- a random key material file `~/.config/webull/key_material` (`0600`, created on first use)
- an optional environment variable `WEBULL_KEY_SALT` (handy for CI)
- a machine fingerprint (hostname + user)

This fallback protects against casual disclosure (accidental commits, backup
leaks, other local users). It is not a substitute for the OS keychain and does
not defend against an attacker with full read access to your home directory and
process environment. The keychain is always preferred when available.

Non-sensitive data (auth method, region, env, token expiry) is stored in
`~/.config/webull/profiles.json` (macOS/Linux) or
`%USERPROFILE%\.config\webull\profiles.json` (Windows) with `0600` permissions.

### Logout

```bash
webull auth logout
```

Removes the stored credentials from the system keychain and clears the local
config record.

### Environment Selection

Use `--env` to choose the environment:

```bash
webull auth login --env prod            # production (default)
webull auth login --env sandbox         # sandbox
```

`sandbox` and `uat` are accepted as equivalent aliases for the non-production
environment. Endpoint hosts for every region and environment are built into the
CLI — you do not need to specify them.

### Environment Variables (AK/SK)

For CI/automation, `WEBULL_APP_KEY` / `WEBULL_APP_SECRET` override the stored
App Key / App Secret at request time. You still need to run `auth login` once
with `--method credentials` so the CLI knows the region and environment:

```bash
webull auth login --method credentials --region us   # establishes region/env

export WEBULL_APP_KEY=your-app-key
export WEBULL_APP_SECRET=your-app-secret
webull account list                                   # uses env-var credentials
```

### Credential Precedence

For AK/SK requests, the App Key / App Secret are resolved in this order:

1. Environment variables (`WEBULL_APP_KEY` + `WEBULL_APP_SECRET`)
2. Stored credentials from the system keychain

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
| `webull order option strategy` | Place an option order via JSON (multi-leg strategies: US only) |
| `webull order futures submit` | Place a futures order |
| `webull order futures replace` | Modify a futures order |
| `webull order crypto submit` | Place a crypto order |
| `webull order event submit` | Place an event contract order |
| `webull order event replace` | Modify an event contract order |
| `webull order cancel` | Cancel an order (any asset) |
| `webull order open` | List open orders |
| `webull order detail` | Get order details |
| `webull order history` | Order history |
| `webull order executions` | List order execution records within a date range (HK) |

### Market Data (`webull data`)

Data is organized by asset class under `webull data <asset>`:

| Command | Description |
|---------|-------------|
| `webull data stock snapshot` | Real-time stock snapshot |
| `webull data stock batch-bars` | Historical candlestick bars (one or more symbols) |
| `webull data stock quotes` | Depth quotes (bid/ask) |
| `webull data stock tick` | Tick-by-tick trades |
| `webull data stock footprint` | Large order footprint |
| `webull data stock noii-bars` | NOII (Net Order Imbalance) bars |
| `webull data stock noii-snapshot` | NOII snapshot |
| `webull data market-sentiment` | US market fear & greed sentiment index (US only) |
| `webull data stock unusual-activities` | US stock unusual activity alerts (US only) |
| `webull data stock insider-transactions` | US insider transactions (US only) |
| `webull data stock insider-holdings-overview` | US insider ownership overview (US only) |
| `webull data stock economic-calendar` | US macroeconomic calendar (US only) |
| `webull data 13f search` | Search US institutions (US only) |
| `webull data 13f holdings` | Institution 13F holdings (US only) |
| `webull data 13f transactions` | Institution holding changes (US only) |
| `webull data news search` | Search US market news by keyword (US only) |
| `webull data news detail` | Get full US news content by ID (US only) |
| `webull data news stock-news` | US stock-related news list (US only) |
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

### Community (`webull community`) — US only

| Command | Description |
|---------|-------------|
| `webull community post create` | Create a plain-text post (optional `--symbol` + `--category`) |
| `webull community comment create` | Create a comment on a post |
| `webull community hot-stocks` | List stocks ranked by community activity |
| `webull community post by-symbol` | List posts associated with a symbol |
| `webull community post get` | Get post detail |
| `webull community my-posts` | List posts created by the current user |

### Account

| Command | Description |
|---------|-------------|
| `webull account list` | List trading accounts |
| `webull account balance` | Account balance |
| `webull account positions` | Account positions |
| `webull account position-detail` | Position details by contract (JP) |

### Authentication

| Command | Description |
|---------|-------------|
| `webull auth login` | Authenticate (OAuth by default, or AK/SK with `--method credentials`) |
| `webull auth logout` | Clear local credentials |
| `webull auth status` | Check auth status |

### Utilities

| Command | Description |
|---------|-------------|
| `webull doctor` | Diagnose config and connectivity |
| `webull version` | Print version |

Every command supports `--help` for full flag documentation:

```bash
webull order stock submit --help    # See all flags for stock submit
webull data stock batch-bars --help # See all flags for bars query
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

## Migration Notes (v1.0.8)

This release moves the commands to the latest API paths and replaces the old paging flags with `--pagination-key`, without a compatibility layer. Scripts that still pass a removed flag fail with `unknown flag`.

### Paging

Use `--pagination-key` with the `next pagination_key` value from the previous response. The last page has no key.

| Command | Removed flags |
|---------|---------------|
| `webull account position-detail` | `--last-id`, `--page-size` |
| `webull data fundamentals fund-dividends` | `--page-index`, `--page-size` |
| `webull data screener market-sectors`, `market-sectors-detail` | `--page-index`, `--page-size` |
| `webull instrument stock`, `instrument crypto`, `instrument event` | `--last-instrument-id`, `--page-size` |
| `webull instrument event-series` | `--last-series-id`, `--page-size` |
| `webull order executions` | `--last-execution-id`, `--page-size` |

`webull data screener gainers-losers`, `actives`, `high-dividend` and `52whl` lost `--page-index` and `--page-size` and have no replacement: these endpoints return a single result set without paging.

The number of rows per page is now decided by the server default, not by `--page-size`.

### Other changes

| Before | Now |
|--------|-----|
| `webull data stock bars` | Removed. Use `webull data stock batch-bars --symbol AAPL` (comma-separated symbols are accepted). |
| `webull instrument option --start-date` / `--end-date` | `--start-date` is now the exact expiration date and `--end-date` is the expiration date lower bound (inclusive). Check any script that used them as a range. |
| `webull profile list / switch / logout / set-endpoint` | Removed. The CLI keeps a single authentication record; each `auth login` overwrites it. Use `webull auth logout` to remove it. |
| Global `--profile`, `--env`, `--region` | Removed from other commands. `webull auth login` keeps `--env` and `--region`. |
| `webull auth login --name`, `--api-endpoint`, `--profile` | Removed. Endpoints for every region are built in. Use `--method oauth` (default) or `--method credentials`. |

New options (no action needed): `--start-time` / `--end-time` (milliseconds) on the `data option|futures|crypto|event bars` and `data stock batch-bars` commands, `--sub-category` on `instrument stock`, `--body` / `--file` on the `order stock|futures|event replace` commands, and `order futures preview` / `order event preview`.

## Safety

- **Order confirmation**: `order stock submit`, `order cancel`, and `order stock replace` prompt for confirmation in interactive mode. Use `-y` to skip.
- **Dry run**: `webull order stock submit --dry-run` prints the request payload without submitting.
- **Credentials**: OAuth tokens and AK/SK secrets are stored in the system keychain, or in an AES-256-GCM encrypted file when the keychain is unavailable — never in plaintext. Non-sensitive config uses `0600` file permissions.

## Global Flags

| Flag | Description |
|------|-------------|
| `--csv` | Output in CSV format |
| `--quiet` | Suppress non-data output |
| `--verbose` | Show request summaries on stderr |
| `--debug` | Show full request/response on stderr |
| `--timeout` | HTTP timeout in seconds (default: 30) |
| `-y, --yes` | Skip confirmation prompts |

`auth login` takes its own `--method`, `--region`, and `--env` flags; see
[Authentication](#authentication) above.

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
