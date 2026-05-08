"""Stripe webhook handler (to be completed in Phase 4)."""
from fastapi import APIRouter, Request

webhook_router = APIRouter(prefix="/webhooks")


@webhook_router.post("/stripe")
async def stripe_webhook(request: Request):
    # TODO: Implement in Phase 4
    return {"received": True}
