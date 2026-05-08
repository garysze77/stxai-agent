"""LLM routing: Puter as primary (OpenAI-compatible endpoint).

Puter uses auth tokens (not traditional API keys).
Sign in at puter.com → get your token → set PUTER_API_KEY in .env.

Model names use provider/model format:
  anthropic/claude-sonnet-4-5, openai/gpt-5-nano, etc.
"""
from langchain_openai import ChatOpenAI
from config import settings


def get_llm_client() -> ChatOpenAI:
    """Puter AI — OpenAI-compatible endpoint with auth token."""
    return ChatOpenAI(
        base_url="https://api.puter.com/puterai/openai/v1/",
        api_key=settings.puter_api_key or "anonymous",
        model="openai/gpt-5-nano",
        max_tokens=2048,
        timeout=60,
        max_retries=2,
    )


def get_fallback_client() -> ChatOpenAI | None:
    """Fallback LLM — only if configured."""
    if not settings.minimax_api_key:
        return None
    return ChatOpenAI(
        base_url="https://api.minimax.chat/v1",
        api_key=settings.minimax_api_key,
        model="MiniMax-M2.5",
        max_tokens=2048,
        timeout=60,
        max_retries=2,
    )
