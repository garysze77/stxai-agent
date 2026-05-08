from fastapi import Request, HTTPException
from models.firebase import get_user_by_api_key, increment_usage, get_user_quota
from datetime import datetime, timezone
import re

# ── Stock query guardrail ──

_STOCK_KEYWORDS = [
    # English — stocks & markets
    "stock", "ticker", "price", "market", "trade", "share", "etf", "index",
    "bull", "bear", "rally", "dip", "correction", "dividend", "yield",
    "nasdaq", "nyse", "hkex", "dow", "sp500", "spx", "hsi", "hang seng",
    "rsi", "macd", "bollinger", "ma50", "ma200", "sma", "ema",
    "pe ratio", "eps", "roe", "roa", "market cap", "volume", "support", "resistance",
    "earnings", "revenue", "profit", "margin", "growth", "valuation",
    "technical", "fundamental", "analyst", "consensus", "target",
    # English — economics & macro
    "interest rate", "inflation", "fed", "fomc", "gdp", "cpi", "ppi",
    "bond", "treasury", "yield curve", "recession", "stimulus",
    "sector", "industry", "ipo", "merger", "acquisition", "spin-off",
    # Chinese — stocks & markets
    "股票", "股價", "股市", "分析", "技術", "基本面", "財報", "業績",
    "牛市", "熊市", "港股", "美股", "大市", "指數", "恒指", "恒生", "道指", "納指",
    "現價", "走勢", "圖表", "新聞", "交易", "買入", "賣出", "持倉",
    "市盈率", "市值", "股息", "回報", "風險",
    # Chinese — economics
    "加息", "減息", "通脹", "利率", "經濟", "GDP", "聯儲局", "央行",
    "板塊", "行業", "上市", "收購", "合併",
]

_TICKER_PATTERNS = [
    re.compile(r'\$[A-Z]{1,5}\b'),                     # $AAPL
    re.compile(r'\b[A-Z]{1,5}\b'),                      # AAPL (loose, guarded by keyword check)
    re.compile(r'\b\d{4,5}\b'),                          # HK ticker: 0700, 9988
    re.compile(r'\b[A-Z0-9]{1,5}\.(HK|US|SS|SZ)\b', re.IGNORECASE),  # 0700.HK
    re.compile(r'(?:ticker|股票代碼|stock code)[:\s]*([A-Z0-9]{1,5})', re.IGNORECASE),
]

# Common non-stock queries — patterns that indicate a non-finance question
_NON_STOCK_PATTERNS = [
    re.compile(r'^(hi|hello|hey|good morning|good evening|what\'?s up)[\s!.,]*$', re.IGNORECASE),
    re.compile(r'^(who|what|where|when|why|how) (are|is|were|was) (you|this|that)\b', re.IGNORECASE),
    re.compile(r'^(thanks?|thank you|ok|okay|bye|goodbye|no|yes|nope|yep)[\s!.,]*$', re.IGNORECASE),
    re.compile(r'^(天氣|今日點|聽日點|你好|早晨|晚安|拜拜|多謝|唔該|係|唔係|好|唔好)[\s!.,]*$'),
    re.compile(r'\b(weather|recipe|cook|movie|song|music|game|sport|football|basketball|porn|nsfw)\b', re.IGNORECASE),
]


def is_stock_query(message: str) -> bool:
    """Check if a message is stock/finance related. Zero-cost pre-filter."""
    if not message or len(message.strip()) < 2:
        return False

    msg = message.strip()
    msg_lower = msg.lower()

    # 1. Reject obvious non-stock queries
    for pat in _NON_STOCK_PATTERNS:
        if pat.search(msg):
            return False

    # 2. Check ticker patterns (strongest signal)
    for pat in _TICKER_PATTERNS:
        if pat.search(msg):
            return True

    # 3. Check stock keywords
    for kw in _STOCK_KEYWORDS:
        if kw in msg_lower:
            return True

    # 4. Allow if message is long enough (likely a substantive question)
    if len(msg) > 30:
        return True

    return False


def validate_stock_query(message: str) -> None:
    """Raise HTTPException if the message is not stock-related."""
    if not is_stock_query(message):
        raise HTTPException(
            status_code=400,
            detail="This AI agent only handles stock market analysis. "
                   "Please ask about stocks, tickers, or financial markets. "
                   "Examples: 'AAPL analysis', '港股 0700 現價', 'What's the outlook for TSLA?'"
        )


# ── API key verification ──

async def verify_api_key(request: Request) -> dict:
    """Extract and validate API key from Authorization header."""
    auth = request.headers.get("Authorization", "")
    if not auth.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing API key")

    api_key = auth.removeprefix("Bearer ").strip()
    if not api_key:
        raise HTTPException(status_code=401, detail="Missing API key")

    user = await get_user_by_api_key(api_key)
    if user is None:
        raise HTTPException(status_code=401, detail="Invalid API key")

    # Check usage quota
    today_count = await increment_usage(api_key, user["id"])
    quota = await get_user_quota(user["id"])

    if today_count > quota:
        raise HTTPException(status_code=429, detail="Daily quota exceeded")

    return user
