"""Real market data tools using yfinance (US + HK stocks)."""
import yfinance as yf


def _to_hk_ticker(ticker: str) -> str:
    """Convert HK ticker format to yfinance format (add .HK suffix)."""
    ticker = ticker.upper().strip()
    if ticker.endswith(".HK"):
        return ticker
    # Pad to 4 digits for HK stocks
    if ticker.isdigit():
        return f"{int(ticker):04d}.HK"
    return ticker


def _get_ticker(ticker: str, market: str = "us") -> yf.Ticker:
    if market == "hk":
        return yf.Ticker(_to_hk_ticker(ticker))
    return yf.Ticker(ticker)


def get_stock_price(ticker: str, market: str = "us") -> dict:
    """Get the current stock price and basic info."""
    try:
        t = _get_ticker(ticker, market)
        info = t.info
        price = info.get("currentPrice") or info.get("regularMarketPrice") or info.get("previousClose")
        return {
            "ticker": ticker.upper(),
            "market": market,
            "name": info.get("shortName") or info.get("longName", ticker),
            "price": price,
            "currency": info.get("currency", "USD" if market == "us" else "HKD"),
            "change_percent": info.get("regularMarketChangePercent"),
            "previous_close": info.get("previousClose"),
            "volume": info.get("volume"),
            "market_cap": info.get("marketCap"),
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "market": market, "error": str(e), "source": "yfinance"}


def get_historical_data(ticker: str, market: str = "us", period: str = "6mo") -> dict:
    """Get historical OHLCV data."""
    try:
        t = _get_ticker(ticker, market)
        df = t.history(period=period)
        if df.empty:
            return {"ticker": ticker, "error": "No data", "source": "yfinance"}

        latest = df.iloc[-1]
        return {
            "ticker": ticker.upper(),
            "market": market,
            "period": period,
            "latest_close": float(latest["Close"]),
            "high_52w": float(df["Close"].rolling(252).max().iloc[-1]) if len(df) > 252 else float(df["High"].max()),
            "low_52w": float(df["Close"].rolling(252).min().iloc[-1]) if len(df) > 252 else float(df["Low"].min()),
            "avg_volume": float(df["Volume"].tail(20).mean()),
            "data_points": len(df),
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "error": str(e), "source": "yfinance"}


def scan_market(market: str = "us", criteria: str = "") -> dict:
    """Basic market scanner."""
    popular = {
        "us": ["AAPL", "MSFT", "GOOGL", "AMZN", "NVDA", "META", "TSLA", "JPM", "V", "JNJ"],
        "hk": ["0700", "9988", "0941", "2318", "0388", "0005", "1299", "0011", "2388", "1810"],
    }
    tickers = popular.get(market, popular["us"])

    results = []
    for ticker in tickers:
        try:
            info = _get_ticker(ticker, market).info
            results.append({
                "ticker": ticker,
                "name": info.get("shortName", ticker),
                "price": info.get("currentPrice") or info.get("regularMarketPrice"),
                "change_percent": info.get("regularMarketChangePercent"),
                "volume": info.get("volume"),
            })
        except Exception:
            results.append({"ticker": ticker, "error": "fetch failed"})

    return {
        "market": market,
        "criteria": criteria or "popular stocks",
        "results": results,
        "source": "yfinance",
    }
