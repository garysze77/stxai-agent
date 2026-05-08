import firebase_admin
from firebase_admin import credentials, firestore
from config import settings

_db: firestore.Client | None = None


def init_firebase():
    global _db
    if _db is not None:
        return _db

    cred_path = settings.google_application_credentials
    if cred_path:
        cred = credentials.Certificate(cred_path)
    else:
        cred = credentials.ApplicationDefault()

    firebase_admin.initialize_app(
        cred, {"projectId": settings.firebase_project_id}
    )
    _db = firestore.client()
    return _db


def get_db() -> firestore.Client:
    if _db is None:
        return init_firebase()
    return _db


# ── User operations ──

async def get_user_by_api_key(api_key: str) -> dict | None:
    db = get_db()
    try:
        docs = (
            db.collection("api_keys")
            .where(filter=("key", "==", api_key))
            .where(filter=("is_active", "==", True))
            .limit(1)
            .stream()
        )
    except Exception as e:
        raise RuntimeError(f"Firestore query failed: {e}")

    for doc in docs:
        key_data = doc.to_dict()
        user_doc = db.collection("users").document(key_data["user_id"]).get()
        if user_doc.exists:
            return {**user_doc.to_dict(), "id": user_doc.id, "key_id": doc.id}
    return None


async def increment_usage(api_key: str, user_id: str) -> int:
    from datetime import datetime, timezone

    db = get_db()
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    usage_id = f"{user_id}:{today}"

    doc_ref = db.collection("usage").document(usage_id)
    doc = doc_ref.get()

    if doc.exists:
        new_count = doc.to_dict().get("count", 0) + 1
        doc_ref.update({"count": new_count})
    else:
        new_count = 1
        doc_ref.set({"user_id": user_id, "date": today, "count": 1})

    return new_count


async def get_user_quota(user_id: str) -> int:
    db = get_db()
    user_doc = db.collection("users").document(user_id).get()
    if not user_doc.exists:
        return 0
    tier = user_doc.to_dict().get("subscription_tier", "free")
    return {
        "free": settings.quota_free,
        "pro": settings.quota_pro,
        "premium": settings.quota_premium,
    }.get(tier, 0)
