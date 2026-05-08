"""Financial analysis tools for the STX AI agent.

Uses real data from yfinance.
"""
from langchain_core.tools import tool
from tools.market_data import get_stock_price as _real_price, scan_market as _real_scan
from tools.technical import get_technical_indicators as _real_ta
from tools.fundamental import (
    get_analyst_ratings as _real_analyst,
    get_fundamental_data as _real_fundamental,
    get_institutional_holders as _real_institutional,
)


@tool
def get_stock_price(ticker: str, market: str = "us") -> dict:
    """Get current stock price and basic info.

    Args:
        ticker: Stock ticker (e.g., AAPL, MSFT for US; 0700, 9988 for HK)
        market: "us" for US stocks, "hk" for Hong Kong stocks
    """
    result = _real_price(ticker, market)
    if "error" in result:
        return _mock_price(ticker, market)
    return result


@tool
def get_technical_indicators(ticker: str, market: str = "us") -> dict:
    """Get technical analysis indicators: RSI(14), MACD, MA(50/200), Bollinger Bands.

    Args:
        ticker: Stock ticker symbol
        market: "us" or "hk"
    """
    result = _real_ta(ticker, market)
    if "error" in result:
        return _mock_ta(ticker, market)
    return result


@tool
def get_analyst_ratings(ticker: str, market: str = "us") -> dict:
    """Get Wall Street analyst consensus rating, price targets, and number of analysts covering the stock.

    Args:
        ticker: Stock ticker symbol
        market: "us" or "hk"
    """
    result = _real_analyst(ticker, market)
    if "error" in result:
        return {"ticker": ticker.upper(), "note": "Analyst data unavailable"}
    return result


@tool
def get_fundamental_data(ticker: str, market: str = "us") -> dict:
    """Get fundamental financial metrics: P/E, EPS, revenue growth, profit margins, ROE, debt/equity, dividends, market cap.

    Args:
        ticker: Stock ticker symbol
        market: "us" or "hk"
    """
    result = _real_fundamental(ticker, market)
    if "error" in result:
        return {"ticker": ticker.upper(), "note": "Fundamental data unavailable"}
    return result


@tool
def get_institutional_holders(ticker: str, market: str = "us") -> dict:
    """Get institutional ownership percentage and top institutional holders (big funds, banks).

    Args:
        ticker: Stock ticker symbol
        market: "us" or "hk"
    """
    result = _real_institutional(ticker, market)
    if "error" in result:
        return {"ticker": ticker.upper(), "note": "Institutional data unavailable"}
    return result


@tool
def get_news(ticker: str, limit: int = 5) -> dict:
    """Get recent news articles for a stock.

    Args:
        ticker: Stock ticker symbol
        limit: Number of articles (max 10)
    """
    try:
        import yfinance as yf
        t = yf.Ticker(ticker)
        news_list = t.news[:limit] if t.news else []
        articles = []
        for n in news_list:
            articles.append({
                "title": n.get("title", ""),
                "publisher": n.get("publisher", ""),
                "link": n.get("link", ""),
                "published": n.get("providerPublishTime", ""),
            })
        return {"ticker": ticker.upper(), "articles": articles, "source": "yfinance"}
    except Exception:
        return _mock_news(ticker, limit)


@tool
def scan_market(market: str = "us", criteria: str = "") -> dict:
    """Scan market for stocks matching criteria (e.g., 'popular stocks', 'top gainers').

    Args:
        market: "us" or "hk"
        criteria: What to scan for
    """
    result = _real_scan(market, criteria)
    if not result.get("results"):
        return _mock_scan(market, criteria)
    return result


# ── Mock fallbacks ──


def _mock_price(ticker: str, market: str) -> dict:
    return {
        "ticker": ticker.upper(), "market": market,
        "name": ticker.upper(), "price": 150.00,
        "currency": "USD" if market == "us" else "HKD",
        "change_percent": 1.5, "volume": 50_000_000,
        "source": "mock",
    }


def _mock_ta(ticker: str, market: str) -> dict:
    return {
        "ticker": ticker.upper(), "market": market,
        "price": 150.00, "rsi_14": 55.0,
        "macd": 1.2, "macd_signal": 0.8, "macd_histogram": 0.4,
        "ma_50": 145.0, "ma_200": 138.0,
        "bollinger_upper": 160.0, "bollinger_middle": 150.0, "bollinger_lower": 140.0,
        "volume_avg_20d": 50_000_000, "source": "mock",
    }


def _mock_news(ticker: str, limit: int) -> dict:
    return {
        "ticker": ticker.upper(),
        "articles": [{"title": f"Latest {ticker.upper()} update", "publisher": "Mock News"}],
        "source": "mock",
    }


def _mock_scan(market: str, criteria: str) -> dict:
    return {
        "market": market, "criteria": criteria, "results": [],
        "note": "Scanner data unavailable", "source": "mock",
    }


ALL_TOOLS = [
    get_stock_price,
    get_technical_indicators,
    get_analyst_ratings,
    get_fundamental_data,
    get_institutional_holders,
    get_news,
    scan_market,
]
TOOLS_BY_NAME = {t.name: t for t in ALL_TOOLS}
