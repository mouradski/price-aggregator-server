# price-aggregator-server

A lightweight, real-time crypto price aggregator written in Go. It connects to
**80+ exchanges** (WebSocket-first, with REST as a startup seed and fallback),
normalises every ticker into a single format, and broadcasts the unified stream
over a WebSocket endpoint.

It tracks the **63 Flare [FTSO](https://dev.flare.network/ftso/) feed assets** by
default, making it suitable as a price source for an FTSO data provider — but the
asset and exchange lists are fully configurable.

## Features

- **80+ exchanges** out of the box (Binance, Coinbase, Kraken, OKX, Bybit,
  KuCoin, Gate.io, Bitget, MEXC, HTX, Crypto.com, Hyperliquid, Deribit, … ).
- **WebSocket-first.** Where an exchange offers a public WS ticker feed it is used
  as the primary source.
- **Hybrid WS + REST.** For hybrid exchanges, REST seeds prices at startup and
  acts as a fallback: it pauses while the WS feed is live and automatically
  resumes if the WS goes silent.
- **Unified ticker schema** broadcast to all clients on `ws://<host>:<port>/ticker`.
- **Tiny footprint:** a single static Go binary, one dependency
  (`gorilla/websocket`), low memory/CPU, instant startup.

## Quick start

```bash
cp .env.example .env        # adjust ASSETS / EXCHANGE if needed
go run ./cmd/aggregator
# ticker stream: ws://localhost:8090/ticker
```

### Docker

```bash
docker compose up -d --build
```

## Configuration (`.env`)

| Variable           | Description                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `ASSETS`           | Comma-separated base assets to track (default: the 63 FTSO feeds). |
| `EXCHANGE`         | Whitelist of enabled exchanges. `all` enables every exchange.      |
| `EXCHANGE_EXCLUDED`| Optional blacklist (takes priority over the whitelist).            |
| `AGGREGATOR_PORT`  | Port for the `/ticker` WebSocket server (default `8090`).          |
| `MESSAGE_TIMEOUT`  | Seconds without a WS message before reconnecting (default `30`).   |

An exchange must be in the `EXCHANGE` whitelist (or `EXCHANGE=all`) to run.

## Output

Each ticker is broadcast as JSON:

```json
{"lastPrice":73785.7,"exchange":"binance","base":"BTC","quote":"USDT","timestamp":1780046281167,"source":"WS","h24Volume":1632195538.84}
```

`source` is `WS` or `REST`. Some exchanges also emit a `<name>-ask` variant
carrying the mid of best bid/ask.

## Architecture

```
cmd/aggregator        entrypoint: load config, start runners + /ticker server
internal/client       WS runner, REST runner, hybrid runner, decompress, helpers
internal/exchanges/*  one package per exchange (WsExchange and/or RestExchange)
internal/registry     allWs() / allRest() / allHybrid() — the enabled exchange lists
internal/server       /ticker WebSocket broadcast server
internal/service      ticker fan-out
internal/symbol       pair (base/quote) parsing
internal/config       .env + environment loading
internal/jsonutil     tolerant Float/String JSON types for inconsistent APIs
```

## Adding an exchange

1. Create `internal/exchanges/<name>/<name>.go` implementing
   `client.WsExchange` (optionally `Ponger` / `Pinger` / `MetadataDecoder` /
   `HeaderProvider`) and/or `client.RestExchange`.
2. Register it in `internal/registry/registry.go` — in `allWs()`, `allRest()`,
   or `allHybrid()` (WS primary + REST fallback sharing the same `Name()`).
3. Add the name to `EXCHANGE` in `.env` (or use `EXCHANGE=all`).

## License

[MIT](LICENSE)
