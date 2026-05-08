from fastapi import APIRouter, Depends, Request, HTTPException
from api.schemas import (
    ChatRequest,
    ChatResponse,
    AnalyzeRequest,
    AnalyzeResponse,
    ScanRequest,
    NewsResponse,
    MarketSummaryResponse,
    ErrorResponse,
)
from api.middleware import verify_api_key

router = APIRouter(prefix="/api/v1")


@router.get("/health")
async def health():
    return {"status": "ok", "service": "stxai-cloud"}


@router.post("/chat", response_model=ChatResponse)
async def chat(body: ChatRequest, user: dict = Depends(verify_api_key)):
    # TODO: Wire to LangGraph agent in Phase 2
    return ChatResponse(
        reply=f"[STX AI] I received your message about '{body.message}'. "
        f"Analysis engine is being built. (tier: {user.get('subscription_tier', 'free')})",
        session_id=body.session_id or "session-placeholder",
    )


@router.get("/analyze/{ticker}", response_model=AnalyzeResponse)
async def analyze(ticker: str, user: dict = Depends(verify_api_key)):
    # TODO: Wire to market data tools in Phase 3
    return AnalyzeResponse(
        ticker=ticker.upper(),
        name="TBD",
        summary="Analysis engine coming soon.",
    )


@router.post("/scan")
async def scan(body: ScanRequest, user: dict = Depends(verify_api_key)):
    # Only Pro+ can scan
    tier = user.get("subscription_tier", "free")
    if tier == "free":
        raise HTTPException(status_code=403, detail="Upgrade to Pro for market scanning")
    return {"results": [], "message": "Scanner coming in Phase 3"}


@router.get("/news/{ticker}", response_model=NewsResponse)
async def news(ticker: str, user: dict = Depends(verify_api_key)):
    tier = user.get("subscription_tier", "free")
    if tier == "free":
        raise HTTPException(status_code=403, detail="Upgrade to Pro for news analysis")
    return NewsResponse(ticker=ticker.upper(), articles=[])


@router.get("/market/summary", response_model=MarketSummaryResponse)
async def market_summary(user: dict = Depends(verify_api_key)):
    from datetime import datetime, timezone

    return MarketSummaryResponse(updated_at=datetime.now(timezone.utc))
