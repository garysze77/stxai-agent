"""Tool definitions for the financial agent. Stubs return mock data (Phase 3 will implement real data)."""
from langchain_core.tools import tool


@tool
def get_stock_price(ticker: str, market: str = "us") -> dict:
    """Get the current stock price and basic info for a given ticker.

    Args:
        ticker: Stock ticker symbol (e.g., AAPL, 0700.HK)
        market: Market code — "us" for US stocks, "hk" for Hong Kong stocks
    """
    # TODO Phase 3: yfinance / akshare
    return {
        "ticker": ticker.upper(),
        "market": market,
        "price": 150.00,
        "currency": "USD" if market == "us" else "HKD",
        "change_percent": 1.5,
        "volume": 50_000_000,
        "source": "stub",
    }


@tool
def get_technical_indicators(ticker: str, market: str = "us") -> dict:
    """Get common technical analysis indicators for a stock.

    Args:
        ticker: Stock ticker symbol
        market: Market code — "us" or "hk"
    """
    # TODO Phase 3: pandas-ta / ta-lib
    return {
        "ticker": ticker.upper(),
        "market": market,
        "rsi_14": 55.0,
        "macd": 1.2,
        "macd_signal": 0.8,
        "ma_50": 145.0,
        "ma_200": 138.0,
        "bollinger_upper": 160.0,
        "bollinger_lower": 140.0,
        "source": "stub",
    }


@tool
def get_news(ticker: str, limit: int = 5) -> dict:
    """Get recent news articles related to a stock.

    Args:
        ticker: Stock ticker symbol
        limit: Number of articles to return (max 10)
    """
    # TODO Phase 3: News API / RSS
    return {
        "ticker": ticker.upper(),
        "articles": [
            {"title": f"Latest {ticker.upper()} market update", "source": "stub", "date": "2026-05-08"},
        ],
        "source": "stub",
    }


@tool
def scan_market(market: str = "us", criteria: str = "") -> dict:
    """Scan the market for stocks matching given criteria.

    Args:
        market: Market to scan — "us" or "hk"
        criteria: Natural language description of scan criteria
            (e.g., "high volume gainers", "oversold tech stocks")
    """
    # TODO Phase 3: yfinance scanner
    return {
        "market": market,
        "criteria": criteria,
        "results": [],
        "note": "Market scanner coming in Phase 3",
        "source": "stub",
    }


ALL_TOOLS = [get_stock_price, get_technical_indicators, get_news, scan_market]
TOOLS_BY_NAME = {t.name: t for t in ALL_TOOLS}
