# STX AI — 自主金融 AI Agent

Open-source autonomous AI agent for **US & HK stock market analysis**. Self-hosted Go client + Cloud API backend. Subscription required for API access.

## Architecture

```
stxai CLI / Telegram Bot (Go, open source)
    │
    ▼
STX AI Cloud API (Python FastAPI, Railway)
    │
    ├── Puter AI (primary LLM) → MiniMax M2.7 (fallback)
    ├── yfinance / akshare (real market data)
    └── Firebase Auth + Stripe (subscription + usage tracking)
```

| Component | Stack | Purpose |
|-----------|-------|---------|
| `agent/` | Go | CLI + Telegram Bot, single binary, ~9MB |
| `cloud/` | Python FastAPI | AI agent backend, tools, LLM routing |
| `web/` | Next.js | Landing page + user dashboard |

## Features

- **Comprehensive stock analysis** — technical indicators, fundamentals, analyst consensus, institutional flows
- **US & HK markets** — real-time data via yfinance and akshare
- **CLI + Telegram bot** — same AI, two interfaces
- **Single binary** — no runtime dependencies, works on macOS/Linux/Windows
- **LLM failover** — Puter AI primary, MiniMax M2.7 automatic fallback

## Quick Start

```bash
# 1. Install
curl -fsSL https://raw.githubusercontent.com/garysze77/stxai/main/agent/install.sh | sh

# 2. Configure — get your API key at https://stxai.vercel.app
stxai setup

# 3. Start chatting
stxai chat

# 4. Or launch Telegram bot
stxai start
```

### Chat Commands

```
/analyze AAPL    — Full stock analysis (price, technical, fundamentals, analyst, institutional)
/analyze 0700    — HK stock analysis (Tencent)
/clear           — Clear session history
/help            — Show help
/quit            — Exit
```

## Build from Source

```bash
git clone https://github.com/garysze77/stxai.git
cd stxai/agent
make build        # Local build → bin/stxai
make build-all    # Cross-compile all platforms
make install      # Install to /usr/local/bin
```

Requires: Go 1.22+

## Install Methods

| Method | Command |
|--------|---------|
| Shell script | `curl -fsSL https://raw.githubusercontent.com/garysze77/stxai/main/agent/install.sh \| sh` |
| GitHub Releases | [Download binary](https://github.com/garysze77/stxai/releases) |
| Homebrew | `brew install garysze77/stxai/stxai` (coming soon) |
| Build from source | `make build` |

## Cloud API

```bash
cd cloud
cp .env.example .env
# Edit .env with your credentials
pip install .
uvicorn src.main:app --reload
```

Requires: Python 3.12+, Firebase, Stripe, Puter AI, MiniMax accounts.

## Subscription Plans

| Tier | Monthly (HKD) | Daily Quota | Features |
|------|---------------|-------------|----------|
| Free | $0 | 10 req | Basic quotes, simple Q&A |
| Pro | $98 | 200 req | Technical analysis, news, watchlist |
| Premium | $298 | 1000 req | Advanced scanning, priority queue |

[Subscribe →](https://stxai.vercel.app)

## License

MIT — see [LICENSE](LICENSE)
