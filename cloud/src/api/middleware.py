from fastapi import Request, HTTPException
from models.firebase import get_user_by_api_key, increment_usage, get_user_quota
from datetime import datetime, timezone


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
