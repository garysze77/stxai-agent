"""Analysis memory store — save, cache, and retrieve analyses.

Two layers:
1. In-memory cache (per-ticker, TTL-based) — fast reuse across users
2. Firestore (per-user, per-ticker) — long-term memory for learning
"""
import logging
import time
from datetime import datetime, timezone
from models.firebase import get_db

logger = logging.getLogger(__name__)
COLLECTION = "analyses"
MAX_HISTORY = 5
DEFAULT_CACHE_TTL = 300  # 5 minutes

# In-memory cache: keyed by ticker (shared across all users)
_report_cache: dict[str, dict] = {}


def _now() -> float:
    return time.time()


def _extract_ticker(messages: list) -> str | None:
    """Best-effort ticker extraction from user messages."""
    for msg in messages:
        content = getattr(msg, "content", "") or ""
        if not content:
            continue
        import re
        # Match common ticker patterns: $AAPL, AAPL, analyze 0700, etc.
        patterns = [
            r"\$([A-Z]{1,5})",          # $AAPL
            r"ticker[:\s]+([A-Z0-9]{1,5})",
            r"(?:analyze|analyse|stock|分析)[:\s]+([A-Z0-9]{1,5})",
        ]
        for pat in patterns:
            m = re.search(pat, content, re.IGNORECASE)
            if m:
                return m.group(1).upper()
        # Fallback: first all-caps 1-5 char word after common prefixes
        m = re.search(r'(?:of|for|on)\s+([A-Z]{1,5})\b', content)
        if m:
            return m.group(1)
    return None


def _extract_key_metrics(bull_thesis: str, bear_thesis: str) -> dict:
    """Extract key numbers from theses for structured storage."""
    metrics = {}
    import re

    for text in [bull_thesis, bear_thesis]:
        for label, pattern in [
            ("price_targets", r'\$?(\d{2,4}(?:[.,]\d{1,2})?)\s*(?:price\s*)?target'),
            ("pe_mentioned", r'P/?E\s*(?:ratio\s*)?(?:of\s*)?(\d+[.,]?\d*)'),
        ]:
            m = re.search(pattern, text, re.IGNORECASE)
            if m:
                metrics[label] = m.group(1)
    return metrics


def save_analysis(
    user_id: str,
    ticker: str,
    bull_thesis: str,
    bear_thesis: str,
    final_report: str,
    session_id: str = "",
) -> str:
    """Save a completed analysis to Firestore. Returns the doc id."""
    db = get_db()
    doc_id = f"{user_id}:{ticker}:{datetime.now(timezone.utc).strftime('%Y%m%d-%H%M%S')}"

    doc = {
        "user_id": user_id,
        "ticker": ticker.upper(),
        "created_at": int(datetime.now(timezone.utc).timestamp()),
        "bull_thesis": bull_thesis[:4000],
        "bear_thesis": bear_thesis[:4000],
        "final_report": final_report[:8000],
        "key_metrics": _extract_key_metrics(bull_thesis, bear_thesis),
        "session_id": session_id,
    }
    db.collection(COLLECTION).document(doc_id).set(doc)
    logger.info(f"Saved analysis {doc_id}")
    return doc_id


def get_past_analyses(user_id: str, ticker: str, limit: int = MAX_HISTORY) -> list[dict]:
    """Retrieve past analyses for the same ticker, most recent first."""
    db = get_db()
    docs = (
        db.collection(COLLECTION)
        .where("user_id", "==", user_id)
        .where("ticker", "==", ticker.upper())
        .order_by("created_at", direction="DESCENDING")
        .limit(limit)
        .stream()
    )
    results = []
    for doc in docs:
        results.append(doc.to_dict())
    return results


def build_memory_context(past_analyses: list[dict]) -> str:
    """Format past analyses into a prompt-friendly context block."""
    if not past_analyses:
        return ""

    lines = [
        "## 📚 Previous Analyses for This Stock",
        "The following are past STX AI analyses you have produced. "
        "Reference these to show learning, track thesis accuracy, "
        "and provide continuity. If the stock moved as predicted, "
        "highlight that. If not, explain what changed.",
        "",
    ]
    for i, a in enumerate(past_analyses, 1):
        created = datetime.fromtimestamp(
            a.get("created_at", 0), tz=timezone.utc
        ).strftime("%Y-%m-%d")
        ticker = a.get("ticker", "??")
        lines.append(f"### Past Analysis #{i} — {ticker} ({created})")
        lines.append("")
        # Summarize key parts instead of full theses (keep prompt lean)
        bull = a.get("bull_thesis", "")[:600]
        bear = a.get("bear_thesis", "")[:600]
        if bull:
            lines.append(f"**Bull case (excerpt):** {bull}")
            lines.append("")
        if bear:
            lines.append(f"**Bear case (excerpt):** {bear}")
            lines.append("")
        lines.append("---")
        lines.append("")

    lines.append(
        "**Instruction:** Use these past analyses to inform your current analysis. "
        "Note any predictions that played out or failed. "
        "This shows users you learn and track your own track record."
    )
    return "\n".join(lines)


# ── Per-ticker report cache (in-memory, shared across users) ──


def get_cached_report(ticker: str, ttl: float = DEFAULT_CACHE_TTL) -> dict | None:
    """Return cached report if still fresh. Returns None on miss or expiry."""
    entry = _report_cache.get(ticker.upper())
    if not entry:
        return None
    if _now() - entry["cached_at"] > ttl:
        del _report_cache[ticker.upper()]
        return None
    return entry


def cache_report(
    ticker: str,
    bull_thesis: str,
    bear_thesis: str,
    final_report: str,
    price: float | None,
) -> None:
    """Store a report in the in-memory cache, keyed by ticker."""
    _report_cache[ticker.upper()] = {
        "ticker": ticker.upper(),
        "bull_thesis": bull_thesis,
        "bear_thesis": bear_thesis,
        "final_report": final_report,
        "price": price,
        "cached_at": _now(),
    }
    logger.info(f"Cached report for {ticker.upper()}")


def _build_price_update_note(ticker: str, old_price: float | None, new_price: float | None, cached_at: float) -> str:
    """Build a price-update header for a cached report."""
    from datetime import datetime, timezone
    analyzed_time = datetime.fromtimestamp(cached_at, tz=timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    note = f"> ⚡ **Quick update**: This analysis was originally run {analyzed_time}.\n"
    if old_price and new_price:
        change = new_price - old_price
        pct = (change / old_price) * 100
        direction = "up" if change > 0 else "down"
        note += f"> Price then: ${old_price:.2f} → now: **${new_price:.2f}** ({direction} {abs(pct):.1f}%)\n"
        note += f"> The full multi-agent debate below reflects data at analysis time. "
        note += f"Current price has been updated for reference.\n"
    elif new_price:
        note += f"> Current price: **${new_price:.2f}**\n"
    note += "\n---\n\n"
    return note
