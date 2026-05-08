from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException
from langchain_core.messages import HumanMessage, AIMessage

from api.schemas import (
    ChatRequest, ChatResponse, AnalyzeResponse,
    ScanRequest, NewsResponse, MarketSummaryResponse,
)
from api.middleware import verify_api_key
from agent.graph import simple_agent, deep_agent
from agent.state import AgentState

router = APIRouter(prefix="/api/v1")


def _extract_reply(result: dict) -> str:
    for msg in reversed(result.get("messages", [])):
        if isinstance(msg, AIMessage) and msg.content:
            return msg.content
    return "No response generated."


@router.get("/health")
async def health():
    return {"status": "ok", "service": "stxai-cloud"}


@router.post("/chat", response_model=ChatResponse)
async def chat(body: ChatRequest, user: dict = Depends(verify_api_key)):
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
    )
    try:
        result = await graph.ainvoke(state)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Agent error: {e}")

    return ChatResponse(
        reply=_extract_reply(result),
        session_id=body.session_id or "default",
    )


@router.get("/analyze/{ticker}", response_model=AnalyzeResponse)
async def analyze(ticker: str, user: dict = Depends(verify_api_key)):
    # Always use deep multi-agent for the dedicated analysis endpoint
    state = AgentState(
        messages=[HumanMessage(content=f"Analyze {ticker} in depth. Use the multi-agent debate framework: build bull case, bear case, and synthesize a comprehensive analysis.")],
        user_id=user["id"],
        subscription_tier=user.get("subscription_tier", "free"),
    )
    result = await deep_agent.ainvoke(state)
    return AnalyzeResponse(
        ticker=ticker.upper(),
        name=ticker.upper(),
        summary=_extract_reply(result),
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
