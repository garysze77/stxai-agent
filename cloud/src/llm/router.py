"""LLM routing: Puter (primary) → MiniMax (fallback)."""
from httpx import AsyncClient
from config import settings


async def chat_completion(
    messages: list[dict],
    tools: list[dict] | None = None,
) -> dict:
    """Try Puter first, fall back to MiniMax."""
    error = None

    # Try Puter
    try:
        return await _puter_chat(messages, tools)
    except Exception as e:
        error = e

    # Fallback to MiniMax
    try:
        return await _minimax_chat(messages, tools)
    except Exception as e:
        raise RuntimeError(f"All LLM providers failed. Puter: {error}, MiniMax: {e}")


async def _puter_chat(messages: list[dict], tools: list[dict] | None) -> dict:
    async with AsyncClient() as client:
        body = {"messages": messages, "model": "claude-sonnet-4-5"}
        if tools:
            body["tools"] = tools
        resp = await client.post(
            f"{settings.puter_api_url}/chat",
            json=body,
            headers={"Authorization": f"Bearer {settings.puter_api_key}"} if settings.puter_api_key else {},
            timeout=60,
        )
        resp.raise_for_status()
        return resp.json()


async def _minimax_chat(messages: list[dict], tools: list[dict] | None) -> dict:
    async with AsyncClient() as client:
        body = {
            "model": "MiniMax-Text-01",
            "messages": messages,
        }
        if tools:
            body["tools"] = tools
        resp = await client.post(
            f"{settings.minimax_api_url}/chat/completions",
            json=body,
            headers={"Authorization": f"Bearer {settings.minimax_api_key}"},
            timeout=60,
        )
        resp.raise_for_status()
        return resp.json()
