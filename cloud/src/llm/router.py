"""LLM routing: Puter (primary) → MiniMax (fallback).

Both Puter and MiniMax are OpenAI-compatible, so we use LangChain's ChatOpenAI
with different base_urls and api_keys.
"""
from langchain_openai import ChatOpenAI
from config import settings


def get_llm_client() -> ChatOpenAI:
    return ChatOpenAI(
        base_url="https://api.puter.com/puterai/openai/v1/",
        api_key=settings.puter_api_key or "anonymous",
        model="claude-sonnet-4-5",
        temperature=0.3,
        timeout=60,
        max_retries=2,
    )


def get_fallback_client() -> ChatOpenAI | None:
    if not settings.minimax_api_key:
        return None
    return ChatOpenAI(
        base_url="https://api.minimax.chat/v1",
        api_key=settings.minimax_api_key,
        model="MiniMax-M2.5",
        temperature=0.3,
        timeout=60,
        max_retries=2,
    )
