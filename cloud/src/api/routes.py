from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Request
from langchain_core.messages import HumanMessage, AIMessage
import logging

from api.schemas import (
    ChatRequest, ChatResponse, AnalyzeResponse, SignalData, QuotaInfo,
    ScanRequest, NewsResponse, MarketSummaryResponse,
)
from api.middleware import verify_api_key, validate_stock_query
from agent.graph import simple_agent, deep_agent
from agent.state import AgentState
from memory.store import (
    get_cached_report, cache_report, _build_price_update_note,
)
from tools.market_data import get_stock_price as _fetch_price

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/api/v1")


def _extract_reply(result: dict) -> str:
    for msg in reversed(result.get("messages", [])):
        if isinstance(msg, AIMessage) and msg.content:
            return msg.content
    return "No response generated."


def _get_current_price(ticker: str, market: str = "us") -> float | None:
    """Fetch current price for a ticker. Returns None on failure."""
    try:
        data = _fetch_price(ticker, market)
        return data.get("price")
    except Exception:
        return None


def _extract_signal(result: dict) -> SignalData | None:
    """Extract signal data from graph result."""
    score = result.get("confidence_score", 0)
    bias = result.get("directional_bias", "")
    strength = result.get("signal_strength", "")
    if not bias and not score and not strength:
        return None
    return SignalData(
        directional_bias=bias,
        confidence_score=score,
        signal_strength=strength,
    )


def _get_quota(request: Request) -> QuotaInfo | None:
    """Extract quota info from request state (set by verify_api_key middleware)."""
    limit = getattr(request.state, "quota_limit", None)
    remaining = getattr(request.state, "quota_remaining", None)
    if limit is not None and remaining is not None:
        return QuotaInfo(limit=limit, remaining=remaining)
    return None


@router.get("/health")
async def health():
    return {"status": "ok", "service": "stxai-cloud"}


@router.get("/health/llm")
async def health_llm():
    """Check LLM connectivity by testing Puter and MiniMax."""
    from llm.router import _puter_client, _minimax_client
    from langchain_core.messages import HumanMessage

    result = {"status": "ok", "providers": {}}

    # Test Puter
    try:
        llm = _puter_client()
        resp = await llm.ainvoke([HumanMessage(content="ping")])
        result["providers"]["puter"] = {
            "status": "ok",
            "model": resp.response_metadata.get("model_name", "unknown"),
        }
    except Exception as e:
        result["providers"]["puter"] = {"status": "error", "detail": str(e)[:200]}
        result["status"] = "degraded"

    # Test MiniMax
    mm = _minimax_client()
    if mm:
        try:
            resp = await mm.ainvoke([HumanMessage(content="ping")])
            result["providers"]["minimax"] = {
                "status": "ok",
                "model": resp.response_metadata.get("model_name", "unknown"),
            }
        except Exception as e:
            result["providers"]["minimax"] = {"status": "error", "detail": str(e)[:200]}
            if result["status"] == "ok":
                result["status"] = "degraded"
    else:
        result["providers"]["minimax"] = {"status": "not_configured"}

    if all(p.get("status") == "error" for p in result["providers"].values()):
        result["status"] = "down"

    return result


@router.post("/chat", response_model=ChatResponse)
async def chat(body: ChatRequest, request: Request, user: dict = Depends(verify_api_key)):
    # Reject non-stock queries
    validate_stock_query(body.message)

    # Choose agent: deep_analysis=true → multi-agent debate
    if body.deep_analysis:
        graph = deep_agent
    else:
        graph = simple_agent

    state = AgentState(
        messages=[HumanMessage(content=body.message)],
        user_id=user["id"],
        subscription_tier=user.get("subscription_tier", "free"),
        session_id=body.session_id or "",
        ticker=body.ticker or "",
    )
    try:
        result = await graph.ainvoke(state)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Agent error: {e}")

    return ChatResponse(
        reply=_extract_reply(result),
        session_id=body.session_id or "default",
        signal=_extract_signal(result),
        quota=_get_quota(request),
    )


@router.get("/analyze/{ticker}", response_model=AnalyzeResponse)
async def analyze(ticker: str, request: Request, user: dict = Depends(verify_api_key)):
    ticker = ticker.upper()

    # ── Cache check ──
    cached = get_cached_report(ticker)
    if cached:
        new_price = _get_current_price(ticker)
        note = _build_price_update_note(
            ticker, cached["price"], new_price, cached["cached_at"],
        )
        return AnalyzeResponse(
            ticker=ticker,
            name=ticker,
            price=new_price,
            summary=note + cached["final_report"],
            quota=_get_quota(request),
        )

    # ── Cache miss: run full multi-agent debate ──
    state = AgentState(
        messages=[HumanMessage(content=f"Analyze {ticker} in depth. Use the multi-agent debate framework: build bull case, bear case, and synthesize a comprehensive analysis.")],
        user_id=user["id"],
        subscription_tier=user.get("subscription_tier", "free"),
        ticker=ticker,
    )
    try:
        result = await deep_agent.ainvoke(state)
    except Exception as e:
        logger.error(f"Agent failed for {ticker}: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Agent error: {e}")

    summary = _extract_reply(result)

    # Cache the result for future requests
    price = _get_current_price(ticker)
    bull = result.get("bullish_thesis", "")
    bear = result.get("bearish_thesis", "")
    cache_report(ticker, bull, bear, summary, price)

    return AnalyzeResponse(
        ticker=ticker,
        name=ticker,
        price=price,
        summary=summary,
        signal=_extract_signal(result),
        quota=_get_quota(request),
    )


@router.post("/scan")
async def scan(body: ScanRequest, user: dict = Depends(verify_api_key)):
    tier = user.get("subscription_tier", "free")
    if tier == "free":
        raise HTTPException(status_code=403, detail="Upgrade to Pro for market scanning")

    state = AgentState(
        messages=[HumanMessage(content=f"Scan the {body.market} market. Criteria: {body.criteria or 'none'}")],
        user_id=user["id"],
        subscription_tier=tier,
    )
    result = await simple_agent.ainvoke(state)
    return {"results": _extract_reply(result), "market": body.market}


@router.get("/news/{ticker}", response_model=NewsResponse)
async def news(ticker: str, user: dict = Depends(verify_api_key)):
    tier = user.get("subscription_tier", "free")
    if tier == "free":
        raise HTTPException(status_code=403, detail="Upgrade to Pro for news analysis")

    state = AgentState(
        messages=[HumanMessage(content=f"Get the latest news for {ticker}")],
        user_id=user["id"],
        subscription_tier=tier,
        ticker=ticker.upper(),
    )
    result = await simple_agent.ainvoke(state)
    return NewsResponse(ticker=ticker.upper(), articles=[{"summary": _extract_reply(result)}])


@router.get("/market/summary", response_model=MarketSummaryResponse)
async def market_summary(user: dict = Depends(verify_api_key)):
    state = AgentState(
        messages=[HumanMessage(content="Summarize today's US and Hong Kong markets.")],
        user_id=user["id"],
        subscription_tier=user.get("subscription_tier", "free"),
    )
    result = await simple_agent.ainvoke(state)
    return MarketSummaryResponse(
        updated_at=datetime.now(timezone.utc),
        us_market={"summary": _extract_reply(result)},
    )
