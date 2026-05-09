"""Inspect and clean up test data in Firestore."""
import os
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from models.firebase import init_firebase, get_db

db = init_firebase()

TEST_USER = "test-user-001"
TEST_KEY = "stx-test-key-001"


def show_all():
    """Show all data related to the test user."""
    # User doc
    user_doc = db.collection("users").document(TEST_USER).get()
    print("=" * 50)
    print("USER")
    print("=" * 50)
    if user_doc.exists:
        data = user_doc.to_dict()
        for k, v in data.items():
            print(f"  {k}: {v}")
    else:
        print(f"  ❌ User {TEST_USER} NOT FOUND")

    # API keys
    print("\n" + "=" * 50)
    print("API KEYS")
    print("=" * 50)
    keys = db.collection("api_keys").where("user_id", "==", TEST_USER).stream()
    key_list = list(keys)
    if not key_list:
        print("  (none)")
    for doc in key_list:
        d = doc.to_dict()
        print(f"  [{doc.id}]")
        for k, v in d.items():
            print(f"    {k}: {v}")

    # Usage
    print("\n" + "=" * 50)
    print("USAGE (last 7 days)")
    print("=" * 50)
    from datetime import datetime, timezone, timedelta
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    usage_docs = (
        db.collection("usage")
        .where("user_id", "==", TEST_USER)
        .stream()
    )
    usage_list = list(usage_docs)
    if not usage_list:
        print("  (none)")
    total = 0
    for doc in usage_list:
        d = doc.to_dict()
        date = d.get("date", "?")
        count = d.get("count", 0)
        total += count
        marker = " ← TODAY" if date == today else ""
        print(f"  {date}: {count} requests{marker}")
    print(f"  TOTAL: {total} requests")


def cleanup_extra_keys():
    """Remove all API keys for test user EXCEPT the known test key."""
    keys = db.collection("api_keys").where("user_id", "==", TEST_USER).stream()
    deleted = 0
    for doc in keys:
        if doc.id != TEST_KEY:
            print(f"  Deleting extra key: {doc.id}")
            doc.reference.delete()
            deleted += 1
        else:
            print(f"  Keeping: {doc.id}")
    if deleted == 0:
        print("  No extra keys to delete.")
    return deleted


def reset_usage():
    """Delete all usage records for the test user."""
    usage_docs = (
        db.collection("usage")
        .where("user_id", "==", TEST_USER)
        .stream()
    )
    count = 0
    for doc in usage_docs:
        print(f"  Deleting usage: {doc.id}")
        doc.reference.delete()
        count += 1
    print(f"  Deleted {count} usage records.")


def set_tier(tier: str):
    """Upgrade/downgrade the test user's subscription tier."""
    valid = {"free", "pro", "premium"}
    if tier not in valid:
        print(f"❌ Invalid tier: {tier}. Must be one of: {', '.join(valid)}")
        return
    db.collection("users").document(TEST_USER).update({"subscription_tier": tier})
    # Update the key name to match
    db.collection("api_keys").document(TEST_KEY).update({"name": f"{tier.title()} Test Key"})
    print(f"✅ {TEST_USER} tier → {tier}")
    show_all()


if __name__ == "__main__":
    import argparse
    p = argparse.ArgumentParser(description="Inspect/clean Firestore test data")
    p.add_argument("action", nargs="?", default="show",
                   choices=["show", "clean-keys", "reset-usage", "clean-all", "set-tier"])
    p.add_argument("--tier", default="premium", choices=["free", "pro", "premium"],
                   help="Tier for set-tier action (default: premium)")
    args = p.parse_args()

    if args.action == "show":
        show_all()
    elif args.action == "clean-keys":
        print("Removing extra API keys...")
        cleanup_extra_keys()
    elif args.action == "reset-usage":
        print("Resetting usage...")
        reset_usage()
    elif args.action == "clean-all":
        print("Removing extra API keys...")
        cleanup_extra_keys()
        print("\nResetting usage...")
        reset_usage()
        print("\n✅ Cleanup complete. Test key stx-test-key-001 preserved.")
    elif args.action == "set-tier":
        set_tier(args.tier)
