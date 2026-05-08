"""Technical analysis indicators implemented with pandas/numpy."""
import numpy as np
import pandas as pd
import yfinance as yf
from tools.market_data import _get_ticker


def _compute_rsi(prices: pd.Series, period: int = 14) -> float:
    delta = prices.diff()
    gain = delta.where(delta > 0, 0.0)
    loss = (-delta).where(delta < 0, 0.0)
    avg_gain = gain.ewm(alpha=1 / period, adjust=False).mean()
    avg_loss = loss.ewm(alpha=1 / period, adjust=False).mean()
    rs = avg_gain / avg_loss.replace(0, np.nan)
    rsi = 100 - (100 / (1 + rs))
    return float(rsi.iloc[-1]) if not pd.isna(rsi.iloc[-1]) else 50.0


def _compute_macd(prices: pd.Series) -> dict:
    ema12 = prices.ewm(span=12, adjust=False).mean()
    ema26 = prices.ewm(span=26, adjust=False).mean()
    macd_line = ema12 - ema26
    signal = macd_line.ewm(span=9, adjust=False).mean()
    histogram = macd_line - signal
    return {
        "macd": round(float(macd_line.iloc[-1]), 4),
        "macd_signal": round(float(signal.iloc[-1]), 4),
        "macd_histogram": round(float(histogram.iloc[-1]), 4),
    }


def _compute_bollinger(prices: pd.Series, period: int = 20) -> dict:
    ma = prices.rolling(period).mean()
    std = prices.rolling(period).std()
    return {
        "bollinger_upper": round(float(ma.iloc[-1] + 2 * std.iloc[-1]), 2),
        "bollinger_middle": round(float(ma.iloc[-1]), 2),
        "bollinger_lower": round(float(ma.iloc[-1] - 2 * std.iloc[-1]), 2),
    }


def get_technical_indicators(ticker: str, market: str = "us") -> dict:
    """Get common technical indicators for a stock."""
    try:
        t = _get_ticker(ticker, market)
        df = t.history(period="1y")
        if df.empty:
            return {"ticker": ticker, "error": "No data available"}

        close = df["Close"]
        return {
            "ticker": ticker.upper(),
            "market": market,
            "price": round(float(close.iloc[-1]), 2),
            "rsi_14": round(_compute_rsi(close, 14), 1),
            **_compute_macd(close),
            "ma_50": round(float(close.rolling(50).mean().iloc[-1]), 2),
            "ma_200": round(float(close.rolling(200).mean().iloc[-1]), 2),
            **_compute_bollinger(close),
            "volume_avg_20d": int(df["Volume"].tail(20).mean()),
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "market": market, "error": str(e)}
