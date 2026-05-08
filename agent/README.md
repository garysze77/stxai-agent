# STX AI Agent

Open-source autonomous AI agent client for US & HK stock analysis.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/garysze77/stxai/main/agent/install.sh | sh
```

Or download from [GitHub Releases](https://github.com/garysze77/stxai/releases).

Or build from source:

```bash
git clone https://github.com/garysze77/stxai.git
cd stxai/agent
make build
```

## Quick Start

```bash
# 1. Configure — enter your API key from https://stxai.vercel.app/dashboard
stxai setup

# 2. Start chatting
stxai chat

# 3. Or launch Telegram bot (optional)
stxai start
```

## Commands

| Command | Description |
|---------|-------------|
| `stxai setup` | Configure API key and settings |
| `stxai chat` | Interactive CLI chat mode |
| `stxai start` | Launch Telegram bot |

## Chat Commands

```
/analyze AAPL    — Quick stock analysis with indicators
/clear           — Clear session history
/help            — Show help
/quit            — Exit
```

## Telegram Bot

1. Create a bot with [@BotFather](https://t.me/BotFather)
2. Copy the token
3. Run `stxai setup` and enter the token
4. Run `stxai start`

## Build from Source

```bash
make build        # Local build
make build-all    # Cross-compile all platforms
make install      # Install to /usr/local/bin
```

## Requirements

- Go 1.22+ (build from source only)
- STX AI Cloud API key (subscribe at https://stxai.vercel.app)
