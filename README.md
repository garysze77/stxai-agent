# STX AI — 自主金融 AI Agent

AI-powered US & HK stock market analysis. Subscription required for API access.

## Architecture

```
stxai CLI / Telegram Bot (Go, open source)  →  https://github.com/garysze77/stxai-agent
    │
    ▼
STX AI Cloud API (Python FastAPI, Railway)
    │
    ├── Puter AI (primary LLM) → MiniMax M2.7 (fallback)
    ├── yfinance / akshare (real market data)
    └── Firebase Auth + Stripe (subscription + usage tracking)
```

| Component | Repo | Stack |
|-----------|------|-------|
| Agent Client | [stxai-agent](https://github.com/garysze77/stxai-agent) (Public) | Go, single binary |
| Cloud API | `cloud/` (Private) | Python FastAPI |
| Landing Page | `web/` (Private) | Next.js |

## Quick Start

```bash
# Install the open-source agent client
curl -fsSL https://raw.githubusercontent.com/garysze77/stxai-agent/main/install.sh | sh

# Configure — get your API key at https://stxai.vercel.app
stxai setup

# Start chatting
stxai chat
```

See [stxai-agent](https://github.com/garysze77/stxai-agent) for full agent documentation.

## Cloud API (Private)

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
