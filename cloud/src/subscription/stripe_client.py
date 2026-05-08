"""Stripe integration: Checkout Sessions, Customer management."""
import stripe
from config import settings

stripe.api_key = settings.stripe_secret_key

TIER_PRICES = {
    "pro": settings.stripe_pro_price_id,
    "premium": settings.stripe_premium_price_id,
}


def create_checkout_session(
    user_id: str,
    email: str,
    tier: str,
    success_url: str = "http://localhost:3000/dashboard?success=true",
    cancel_url: str = "http://localhost:3000/pricing?canceled=true",
) -> dict:
    """Create a Stripe Checkout Session for subscription."""
    price_id = TIER_PRICES.get(tier)
    if not price_id:
        raise ValueError(f"Unknown tier: {tier}")

    session = stripe.checkout.Session.create(
        mode="subscription",
        line_items=[{"price": price_id, "quantity": 1}],
        customer_email=email,
        metadata={"user_id": user_id, "tier": tier},
        success_url=success_url,
        cancel_url=cancel_url,
        subscription_data={"metadata": {"user_id": user_id, "tier": tier}},
    )
    return {"url": session.url, "session_id": session.id}
