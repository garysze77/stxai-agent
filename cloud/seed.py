"""Seed Firestore with test data for local development."""
import os
import sys
import uuid

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from config import settings
from models.firebase import init_firebase, get_db

db = init_firebase()


def seed():
    # Create a test user
    user_id = "test-user-001"
    db.collection("users").document(user_id).set({
        "email": "test@stxai.io",
        "name": "Test User",
        "created_at": firestore.SERVER_TIMESTAMP,
        "subscription_tier": "premium",
        "subscription_expiry": None,
        "stripe_customer_id": None,
    })
    print(f"Created user: {user_id}")

    # Create a test API key
    api_key = "stx-test-key-001"
    db.collection("api_keys").document(api_key).set({
        "user_id": user_id,
        "key": api_key,
        "created_at": firestore.SERVER_TIMESTAMP,
        "is_active": True,
        "name": "Dev Test Key",
    })
    print(f"Created API key: {api_key}")

    print("Seed complete. Use this API key for testing:")
    print(f"  Authorization: Bearer {api_key}")


if __name__ == "__main__":
    from firebase_admin import firestore
    seed()
