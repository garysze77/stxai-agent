"""Subscription API routes."""
from fastapi import APIRouter, HTTPException, Request
from api.schemas import SubscribeRequest
from subscription.stripe_client import create_checkout_session
from subscription.webhook import process_webhook

sub_router = APIRouter(prefix="/api/v1", tags=["subscription"])
webhook_router = APIRouter(tags=["webhooks"])


@sub_router.post("/subscribe")
async def create_subscription(body: SubscribeRequest):
    """Create a Stripe Checkout Session and return the URL."""
    try:
        result = create_checkout_session(
            user_id=body.user_id,
            email=body.email,
            tier=body.tier,
        )
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@webhook_router.post("/webhooks/stripe")
async def stripe_webhook(request: Request):
    """Receive Stripe webhook events."""
    payload = await request.body()
    signature = request.headers.get("stripe-signature", "")
    try:
        result = await process_webhook(payload, signature)
        return result
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
