"""Simulate Stripe webhook payloads to test all handlers.
Does NOT touch Firestore or Stripe — validates routing + tier extraction.
"""
import os, sys, json, asyncio
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from config import settings


def build_payloads():
    """Build test payloads using real price IDs from settings."""
    pro_price = settings.stripe_pro_price_id or "price_pro_placeholder"
    prem_price = settings.stripe_premium_price_id or "price_premium_placeholder"
    cs = "cus_test_456"
    sub = "sub_test_789"
    user = "test-user-001"

    return [
        ("checkout.session.completed (normal)", {
            "type": "checkout.session.completed",
            "data": {"object": {
                "id": "cs_test_abc", "customer": cs, "subscription": sub,
                "metadata": {"user_id": user, "tier": "pro"},
            }}
        }),
        ("checkout.session.completed (async, no sub)", {
            "type": "checkout.session.completed",
            "data": {"object": {
                "id": "cs_test_async", "customer": cs, "subscription": None,
                "metadata": {"user_id": user, "tier": "premium"},
            }}
        }),
        ("invoice.paid (subscription_create)", {
            "type": "invoice.paid",
            "data": {"object": {
                "id": "in_create", "customer": cs, "subscription": sub,
                "billing_reason": "subscription_create",
                "lines": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("invoice.paid (subscription_cycle)", {
            "type": "invoice.paid",
            "data": {"object": {
                "id": "in_cycle", "customer": cs, "subscription": sub,
                "billing_reason": "subscription_cycle",
                "lines": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("invoice.paid (subscription_update → upgrade)", {
            "type": "invoice.paid",
            "data": {"object": {
                "id": "in_upgrade", "customer": cs, "subscription": sub,
                "billing_reason": "subscription_update",
                "lines": {"data": [{"price": {"id": prem_price}}]},
            }}
        }),
        ("invoice.payment_succeeded", {
            "type": "invoice.payment_succeeded",
            "data": {"object": {
                "id": "in_success", "customer": cs, "subscription": sub,
                "billing_reason": "subscription_cycle",
                "lines": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("customer.subscription.updated (upgrade)", {
            "type": "customer.subscription.updated",
            "data": {"object": {
                "id": sub, "customer": cs, "status": "active",
                "cancel_at_period_end": False,
                "items": {"data": [{"price": {"id": prem_price}}]},
            }}
        }),
        ("customer.subscription.updated (past_due)", {
            "type": "customer.subscription.updated",
            "data": {"object": {
                "id": sub, "customer": cs, "status": "past_due",
                "cancel_at_period_end": False,
                "items": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("customer.subscription.updated (unpaid → revoke)", {
            "type": "customer.subscription.updated",
            "data": {"object": {
                "id": sub, "customer": cs, "status": "unpaid",
                "cancel_at_period_end": False,
                "items": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("customer.subscription.updated (cancel at period end)", {
            "type": "customer.subscription.updated",
            "data": {"object": {
                "id": sub, "customer": cs, "status": "active",
                "cancel_at_period_end": True,
                "items": {"data": [{"price": {"id": pro_price}}]},
            }}
        }),
        ("customer.subscription.deleted", {
            "type": "customer.subscription.deleted",
            "data": {"object": {
                "id": sub, "customer": cs, "status": "canceled",
            }}
        }),
        ("invoice.payment_failed", {
            "type": "invoice.payment_failed",
            "data": {"object": {
                "id": "in_fail", "customer": cs, "subscription": sub,
                "attempt_count": 2,
            }}
        }),
    ]


def mock_settings():
    """Ensure test settings won't accidentally connect to real services."""
    settings.stripe_webhook_secret = ""
    # Must keep real price IDs for tier extraction tests


def test_tier_extraction():
    """Test price_id → tier mapping uses real configured prices."""
    from subscription.webhook import _get_tier_from_invoice, _get_tier_from_subscription
    from subscription.webhook import PRICE_TO_TIER

    print(f"  Configured prices: {json.dumps(PRICE_TO_TIER, indent=4)}")

    pro_price = settings.stripe_pro_price_id
    prem_price = settings.stripe_premium_price_id

    if pro_price:
        inv = {"lines": {"data": [{"price": {"id": pro_price}}]}}
        tier = _get_tier_from_invoice(inv)
        assert tier == "pro", f"Expected pro, got {tier}"
        print(f"  ✅ {pro_price} → pro")

    if prem_price:
        inv = {"lines": {"data": [{"price": {"id": prem_price}}]}}
        tier = _get_tier_from_invoice(inv)
        assert tier == "premium", f"Expected premium, got {tier}"
        print(f"  ✅ {prem_price} → premium")

    inv_unknown = {"lines": {"data": [{"price": {"id": "unknown_price_xyz"}}]}}
    assert _get_tier_from_invoice(inv_unknown) == "free", "Unknown price should return free"
    print("  ✅ unknown price → free")

    inv_empty = {"lines": {"data": []}}
    assert _get_tier_from_invoice(inv_empty) == "free", "Empty lines should return free"
    print("  ✅ empty lines → free")

    return True


async def test_routing():
    """Test that every payload routes to the correct handler."""
    from subscription.webhook import process_webhook

    payloads = build_payloads()
    errors = []

    for name, payload in payloads:
        event_type = payload["type"]
        try:
            body = json.dumps(payload).encode()
            result = await process_webhook(body, "")
            assert result["event"] == event_type, f"Event mismatch: {result['event']} != {event_type}"
            print(f"  ✅ {name}")
        except Exception as e:
            print(f"  ❌ {name}: {e}")
            errors.append((name, str(e)))

    return errors


async def main():
    mock_settings()

    print("=" * 60)
    print("WEBHOOK EVENT ROUTING TEST")
    print("=" * 60)
    errors = await test_routing()

    print()
    print("=" * 60)
    print("TIER EXTRACTION TEST")
    print("=" * 60)
    try:
        test_tier_extraction()
    except Exception as e:
        print(f"  ❌ Tier extraction failed: {e}")
        errors.append(("tier_extraction", str(e)))

    print()
    if errors:
        print(f"❌ {len(errors)} failure(s):")
        for name, err in errors:
            print(f"   - {name}: {err}")
        sys.exit(1)
    else:
        print(f"✅ All tests passed! (Firestore writes skipped — handlers log instead)")


if __name__ == "__main__":
    asyncio.run(main())
