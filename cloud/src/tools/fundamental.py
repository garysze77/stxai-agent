"""Fundamental analysis tools: analyst ratings, financial statements, institutional holders."""
import pandas as pd
from tools.market_data import _get_ticker


def get_analyst_ratings(ticker: str, market: str = "us") -> dict:
    """Get analyst consensus, price targets, and recommendation trends for a stock."""
    try:
        t = _get_ticker(ticker, market)
        info = t.info

        return {
            "ticker": ticker.upper(),
            "market": market,
            "consensus": info.get("recommendationKey", "N/A"),
            "recommendation_score": info.get("recommendationMean"),
            "num_analysts": info.get("numberOfAnalystOpinions", 0),
            "price_targets": {
                "current": info.get("currentPrice"),
                "high": info.get("targetHighPrice"),
                "low": info.get("targetLowPrice"),
                "mean": info.get("targetMeanPrice"),
                "median": info.get("targetMedianPrice"),
            },
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "market": market, "error": str(e), "source": "yfinance"}


def get_fundamental_data(ticker: str, market: str = "us") -> dict:
    """Get key fundamental metrics: valuation, profitability, financial health, dividends."""
    try:
        t = _get_ticker(ticker, market)
        info = t.info

        return {
            "ticker": ticker.upper(),
            "market": market,
            "valuation": {
                "pe_ratio": info.get("trailingPE"),
                "forward_pe": info.get("forwardPE"),
                "peg_ratio": info.get("pegRatio"),
                "price_to_book": info.get("priceToBook"),
                "price_to_sales": info.get("priceToSalesTrailing12Months"),
            },
            "profitability": {
                "eps": info.get("trailingEps"),
                "revenue_growth": info.get("revenueGrowth"),
                "earnings_growth": info.get("earningsGrowth"),
                "profit_margin": info.get("profitMargins"),
                "roe": info.get("returnOnEquity"),
                "roa": info.get("returnOnAssets"),
            },
            "financial_health": {
                "debt_to_equity": info.get("debtToEquity"),
                "current_ratio": info.get("currentRatio"),
                "free_cashflow": info.get("freeCashflow"),
                "total_cash": info.get("totalCash"),
                "total_debt": info.get("totalDebt"),
            },
            "dividend": {
                "dividend_yield": info.get("dividendYield"),
                "dividend_rate": info.get("dividendRate"),
                "payout_ratio": info.get("payoutRatio"),
            },
            "market_cap": info.get("marketCap"),
            "enterprise_value": info.get("enterpriseValue"),
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "market": market, "error": str(e), "source": "yfinance"}


def get_institutional_holders(ticker: str, market: str = "us") -> dict:
    """Get institutional ownership data and top institutional holders."""
    try:
        t = _get_ticker(ticker, market)
        info = t.info

        holders = []
        try:
            df = t.institutional_holders
            if df is not None and not df.empty:
                for _, row in df.head(10).iterrows():
                    holders.append({
                        "holder": str(row.get("Holder", "")),
                        "shares": int(row.get("Shares", 0)) if pd.notna(row.get("Shares")) else 0,
                        "date_reported": str(row.get("Date Reported", "")),
                        "pct_out": float(row.get("% Out", 0)) if pd.notna(row.get("% Out", 0)) else 0,
                    })
        except Exception:
            pass

        return {
            "ticker": ticker.upper(),
            "market": market,
            "institutional_ownership_pct": info.get("heldPercentInstitutions"),
            "insider_ownership_pct": info.get("heldPercentInsiders"),
            "top_holders": holders,
            "source": "yfinance",
        }
    except Exception as e:
        return {"ticker": ticker, "market": market, "error": str(e), "source": "yfinance"}
