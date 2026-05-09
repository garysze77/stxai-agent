"""Stripe webhook handler — full subscription lifecycle."""
import secrets
import time
import logging
import stripe
from config import settings
from models.firebase import (
    get_db,
    get_user_by_stripe_customer,
    get_user_by_stripe_subscription,
)

logger = logging.getLogger(__name__)
stripe.api_key = settings.stripe_secret_key

# Reverse mapping: price ID → tier name
PRICE_TO_TIER = {
    settings.stripe_pro_price_id: "pro",
    settings.stripe_premium_price_id: "premium",
}


def generate_api_key() -> str:
    return "stx-" + secrets.token_urlsafe(24)


def _get_tier_from_subscription(sub: dict) -> str:
    """Extract tier from a subscription object's line items."""
    items = sub.get("items", {}).get("data", [])
    if items:
        price_id = items[0].get("price", {}).get("id", "")
        return PRICE_TO_TIER.get(price_id, "free")
    return "free"


def _get_tier_from_invoice(invoice: dict) -> str:
    """Extract tier from an invoice's line items."""
    lines = invoice.get("lines", {}).get("data", [])
    if lines:
        price_id = lines[0].get("price", {}).get("id", "")
        return PRICE_TO_TIER.get(price_id, "free")
    return "free"


def _find_user(invoice_or_sub: dict) -> dict | None:
    """Find user by Stripe customer ID or subscription ID."""
    customer_id = invoice_or_sub.get("customer", "")
    subscription_id = invoice_or_sub.get("subscription", "") or invoice_or_sub.get("id", "")

    user = None
    if customer_id:
        user = get_user_by_stripe_customer(customer_id)
    if not user and subscription_id:
        user = get_user_by_stripe_subscription(subscription_id)
    return user


# ── Event handlers ──


def handle_checkout_completed(session: dict):
    """Initial subscription provisioning via Checkout."""
    db = get_db()
    user_id = session.get("metadata", {}).get("user_id", "")
    tier = session.get("metadata", {}).get("tier", "free")
    stripe_customer_id = session.get("customer", "")
    subscription_id = session.get("subscription", "")

    if not user_id:
        logger.error("Checkout session missing user_id in metadata")
        return

    update: dict = {
        "subscription_tier": tier,
        "stripe_customer_id": stripe_customer_id,
    }
    if subscription_id:
        update["stripe_subscription_id"] = subscription_id
    else:
        logger.warning(
            "checkout.session.completed without subscription — "
            "saving customer ID for async payment. Will be completed by invoice.payment_succeeded."
        )

    db.collection("users").document(user_id).update(update)

    # Generate API key if the user doesn't have one
    existing_keys = (
        db.collection("api_keys")
        .where("user_id", "==", user_id)
        .where("is_active", "==", True)
        .limit(1)
        .stream()
    )
    has_key = any(True for _ in existing_keys)

    if not has_key:
        api_key = generate_api_key()
        db.collection("api_keys").document(api_key).set({
            "user_id": user_id,
            "key": api_key,
            "created_at": int(time.time()),
            "is_active": True,
            "name": f"{tier.title()} Subscription Key",
        })
        logger.info(f"Provisioned {tier} for user {user_id}: api_key={api_key}")
    else:
        logger.info(f"User {user_id} already has an active key — updated tier to {tier}")


def handle_invoice_payment_succeeded(invoice: dict):
    """Handle successful payment — initial or recurring."""
    db = get_db()
    billing_reason = invoice.get("billing_reason", "")
    stripe_customer_id = invoice.get("customer", "")
    subscription_id = invoice.get("subscription", "")

    user = _find_user(invoice)
    if not user:
        logger.warning(
            f"No user found for invoice payment: "
            f"customer={stripe_customer_id}, subscription={subscription_id}"
        )
        return

    user_id = user["id"]
    tier = _get_tier_from_invoice(invoice)

    if billing_reason == "subscription_create":
        # Initial payment — ensure user is provisioned
        update = {
            "subscription_tier": tier,
            "stripe_customer_id": stripe_customer_id,
            "stripe_subscription_id": subscription_id,
        }
        db.collection("users").document(user_id).update(update)
        logger.info(f"Initial payment for user {user_id}: tier={tier}")

        # Generate key if needed
        existing = (
            db.collection("api_keys")
            .where("user_id", "==", user_id)
            .where("is_active", "==", True)
            .limit(1)
            .stream()
        )
        if not any(True for _ in existing):
            api_key = generate_api_key()
            db.collection("api_keys").document(api_key).set({
                "user_id": user_id,
                "key": api_key,
                "created_at": int(time.time()),
                "is_active": True,
                "name": f"{tier.title()} Subscription Key",
            })
            logger.info(f"Generated API key for user {user_id} via invoice webhook")

    elif billing_reason == "subscription_cycle":
        # Recurring renewal — confirm subscription still active
        db.collection("users").document(user_id).update({
            "subscription_tier": tier,
            "stripe_subscription_id": subscription_id,
        })
        logger.info(f"Renewal payment for user {user_id}: tier={tier}")

    elif billing_reason == "subscription_update":
        # Plan change with payment (e.g., upgrade)
        db.collection("users").document(user_id).update({
            "subscription_tier": tier,
        })
        logger.info(f"Subscription updated for user {user_id}: tier={tier}")

    else:
        logger.info(f"Invoice payment for user {user_id}: billing_reason={billing_reason}")


def handle_subscription_updated(sub: dict):
    """Handle subscription status/plan changes."""
    db = get_db()
    subscription_id = sub.get("id", "")
    status = sub.get("status", "")
    cancel_at_end = sub.get("cancel_at_period_end", False)

    user = _find_user(sub)
    if not user:
        logger.warning(f"No user found for subscription update: {subscription_id}")
        return

    user_id = user["id"]
    tier = _get_tier_from_subscription(sub)

    update = {"subscription_tier": tier}
    logger.info(
        f"Subscription updated for user {user_id}: "
        f"status={status}, tier={tier}, cancel_at_period_end={cancel_at_end}"
    )

    if status == "past_due":
        # Payment failed but still in retry window — don't revoke yet
        logger.warning(f"Subscription past_due for user {user_id} — waiting for retry")

    elif status in ("unpaid", "canceled"):
        # Payment retries exhausted or canceled — revoke access
        update["subscription_tier"] = "free"
        update["stripe_subscription_id"] = None
        db.collection("users").document(user_id).update(update)
        _deactivate_user_keys(user_id)
        logger.info(f"Revoked access for user {user_id} due to status={status}")
        return

    elif cancel_at_end:
        # Scheduled cancellation — keep access until period end
        logger.info(f"User {user_id} will cancel at period end")

    db.collection("users").document(user_id).update(update)


def handle_subscription_deleted(sub: dict):
    """Subscription ended — revoke access."""
    db = get_db()
    subscription_id = sub.get("id", "")

    user = _find_user(sub)
    if not user:
        logger.warning(f"No user found for subscription deletion: {subscription_id}")
        return

    user_id = user["id"]
    db.collection("users").document(user_id).update({
        "subscription_tier": "free",
        "stripe_subscription_id": None,
    })

    _deactivate_user_keys(user_id)
    logger.info(f"Revoked subscription for user {user_id}")


def handle_invoice_payment_failed(invoice: dict):
    """Payment failed — log and wait for Stripe retry or subscription.deleted."""
    stripe_customer_id = invoice.get("customer", "")
    subscription_id = invoice.get("subscription", "")
    attempt = invoice.get("attempt_count", 1)

    user = _find_user(invoice)
    user_id = user["id"] if user else "unknown"

    logger.warning(
        f"Payment failed for user {user_id}: "
        f"customer={stripe_customer_id}, subscription={subscription_id}, "
        f"attempt={attempt}"
    )
    # Don't revoke yet — Stripe will retry. If all retries fail,
    # customer.subscription.updated (status=unpaid) or
    # customer.subscription.deleted will fire.


# ── Helpers ──


def _deactivate_user_keys(user_id: str):
    """Deactivate all active API keys for a user."""
    db = get_db()
    keys = (
        db.collection("api_keys")
        .where("user_id", "==", user_id)
        .where("is_active", "==", True)
        .stream()
    )
    count = 0
    for doc in keys:
        doc.reference.update({"is_active": False})
        count += 1
    if count:
        logger.info(f"Deactivated {count} API key(s) for user {user_id}")


# ── Main webhook processor ──


async def process_webhook(payload: bytes, signature: str) -> dict:
    """Verify and route a Stripe webhook event."""
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
        import json
        event = json.loads(payload)

    event_type = event["type"]
    event_obj = event["data"]["object"]

    handlers = {
        "checkout.session.completed": lambda: handle_checkout_completed(event_obj),
        "invoice.payment_succeeded": lambda: handle_invoice_payment_succeeded(event_obj),
        "invoice.paid": lambda: handle_invoice_payment_succeeded(event_obj),
        "customer.subscription.updated": lambda: handle_subscription_updated(event_obj),
        "customer.subscription.deleted": lambda: handle_subscription_deleted(event_obj),
        "invoice.payment_failed": lambda: handle_invoice_payment_failed(event_obj),
    }

    handler = handlers.get(event_type)
    if handler:
        try:
            handler()
        except Exception as e:
            logger.error(f"Webhook handler error for {event_type}: {e}", exc_info=True)
    else:
        logger.info(f"Unhandled webhook event: {event_type}")

    return {"status": "ok", "event": event_type}
