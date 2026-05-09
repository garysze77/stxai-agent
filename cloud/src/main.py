from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from api.routes import router
from subscription.routes import sub_router, webhook_router
from models.firebase import init_firebase
from config import settings

app = FastAPI(
    title="STX AI Cloud API",
    description="Financial AI Agent API — stocks, forex, crypto analysis",
    version="0.1.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
    expose_headers=["X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"],
)

# Rate-limit response headers
@app.middleware("http")
async def add_ratelimit_headers(request: Request, call_next):
    response = await call_next(request)
    limit = getattr(request.state, "quota_limit", None)
    remaining = getattr(request.state, "quota_remaining", None)
    if limit is not None:
        response.headers["X-RateLimit-Limit"] = str(limit)
    if remaining is not None:
        response.headers["X-RateLimit-Remaining"] = str(remaining)
    # Reset at midnight UTC
    from datetime import datetime, timezone, timedelta
    tomorrow = datetime.now(timezone.utc).replace(hour=0, minute=0, second=0, microsecond=0) + timedelta(days=1)
    response.headers["X-RateLimit-Reset"] = str(int(tomorrow.timestamp()))
    return response

app.include_router(router)
app.include_router(sub_router)
app.include_router(webhook_router)


@app.on_event("startup")
async def startup():
    init_firebase()


if __name__ == "__main__":
    import os
    import uvicorn

    port = int(os.environ.get("PORT", settings.api_port))
    uvicorn.run(app, host="0.0.0.0", port=port)
