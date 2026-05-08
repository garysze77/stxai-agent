"""Stripe webhook handler."""
import secrets
import time
import logging
import stripe
from config import settings
from models.firebase import get_db

logger = logging.getLogger(__name__)
stripe.api_key = settings.stripe_secret_key


def generate_api_key() -> str:
    return "stx-" + secrets.token_urlsafe(24)


async def handle_checkout_completed(session: dict):
    """Provision subscription: update Firestore, generate API key."""
    db = get_db()
    user_id = session.get("metadata", {}).get("user_id", "")
    tier = session.get("metadata", {}).get("tier", "free")
    customer_email = session.get("customer_email") or session.get("customer_details", {}).get("email", "")
    stripe_customer_id = session.get("customer", "")
    subscription_id = session.get("subscription", "")

    if not user_id:
        logger.error("Checkout session missing user_id in metadata")
        return

    # Update user subscription
    db.collection("users").document(user_id).update({
        "subscription_tier": tier,
        "stripe_customer_id": stripe_customer_id,
        "stripe_subscription_id": subscription_id,
    })

    # Generate API key
    api_key = generate_api_key()
    db.collection("api_keys").document(api_key).set({
        "user_id": user_id,
        "key": api_key,
        "created_at": int(time.time()),
        "is_active": True,
        "name": f"{tier.title()} Subscription Key",
    })

    logger.info(f"Provisioned {tier} subscription for user {user_id}: api_key={api_key}")


async def handle_subscription_deleted(subscription: dict):
    """Revoke subscription: downgrade to free, deactivate API keys."""
    db = get_db()
    user_id = subscription.get("metadata", {}).get("user_id", "")
    if not user_id:
        return

    db.collection("users").document(user_id).update({
        "subscription_tier": "free",
        "stripe_subscription_id": None,
    })

    # Deactivate all API keys for this user
    keys = (
        db.collection("api_keys")
        .where("user_id", "==", user_id)
        .where("is_active", "==", True)
        .stream()
    )
    for doc in keys:
        doc.reference.update({"is_active": False})

    logger.info(f"Revoked subscription for user {user_id}")


async def process_webhook(payload: bytes, signature: str) -> dict:
    """Verify and process a Stripe webhook event."""
    if settings.stripe_webhook_secret:
        try:
            event = stripe.Webhook.construct_event(
                payload, signature, settings.stripe_webhook_secret
            )
        except ValueError:
            raise ValueError("Invalid payload")
        except stripe.error.SignatureVerificationError:
            raise ValueError("Invalid signature")
    else:
        # Development mode: skip signature verification
        import json
        event = json.loads(payload)

    handlers = {
        "checkout.session.completed": handle_checkout_completed,
        "customer.subscription.deleted": handle_subscription_deleted,
    }

    handler = handlers.get(event["type"])
    if handler:
        await handler(event["data"]["object"])

    return {"status": "ok", "event": event["type"]}
