"""LLM routing: Puter (primary) → MiniMax (fallback).

Both use OpenAI-compatible endpoints so we can use langchain_openai.ChatOpenAI for both.
"""
import logging
from langchain_openai import ChatOpenAI
from config import settings

logger = logging.getLogger(__name__)


def _puter_client() -> ChatOpenAI:
    return ChatOpenAI(
        base_url="https://api.puter.com/puterai/openai/v1/",
        api_key=settings.puter_api_key or "anonymous",
        model="openai/gpt-5-nano",
        max_tokens=2048,
        timeout=60,
        max_retries=1,
    )


def _minimax_client() -> ChatOpenAI | None:
    if not settings.minimax_api_key:
        return None
    return ChatOpenAI(
        base_url="https://api.minimax.io/v1",
        api_key=settings.minimax_api_key,
        model="MiniMax-M2.7",
        max_tokens=2048,
        timeout=60,
        max_retries=1,
    )


class LLMRouter:
    """Tries Puter first, falls back to MiniMax automatically."""

    def __init__(self):
        self.primary = _puter_client()
        self.fallback = _minimax_client()

    async def ainvoke(self, messages, tools=None):
        """Invoke LLM with automatic fallback on failure."""
        llm = self.primary
        if tools:
            llm = llm.bind_tools(tools)

        try:
            return await llm.ainvoke(messages), "puter"
        except Exception as e:
            if not self.fallback:
                raise e
            logger.warning(f"Puter failed, falling back to MiniMax: {e}")
            fb_llm = self.fallback
            if tools:
                fb_llm = fb_llm.bind_tools(tools)
            try:
                return await fb_llm.ainvoke(messages), "minimax"
            except Exception as e2:
                logger.error(f"MiniMax also failed: {e2}")
                raise e  # bubble up original Puter error
