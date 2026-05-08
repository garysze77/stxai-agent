from pydantic import BaseModel
from datetime import datetime


class ChatRequest(BaseModel):
    message: str
    ticker: str | None = None
    session_id: str | None = None


class ChatResponse(BaseModel):
    reply: str
    session_id: str
    tokens_used: int | None = None


class AnalyzeRequest(BaseModel):
    ticker: str


class AnalyzeResponse(BaseModel):
    ticker: str
    name: str
    price: float | None = None
    currency: str = "USD"
    summary: str
    technical_signals: dict | None = None


class ScanRequest(BaseModel):
    market: str = "us"  # "us" | "hk"
    criteria: dict | None = None


class NewsResponse(BaseModel):
    ticker: str
    articles: list[dict]


class MarketSummaryResponse(BaseModel):
    us_market: dict | None = None
    hk_market: dict | None = None
    updated_at: datetime


class SubscribeRequest(BaseModel):
    user_id: str
    email: str
    tier: str = "pro"  # "pro" | "premium"


class ErrorResponse(BaseModel):
    detail: str
