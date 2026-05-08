from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    model_config = {"env_file": ".env", "env_file_encoding": "utf-8"}

    # Firebase
    firebase_project_id: str = "stxai-3af70"
    google_application_credentials: str = ""

    # Puter AI
    puter_api_url: str = "https://api.puter.com/v1"
    puter_api_key: str = ""

    # MiniMax (fallback)
    minimax_api_key: str = ""
    minimax_api_url: str = "https://api.minimaxi.chat/v1"

    # Stripe
    stripe_secret_key: str = ""
    stripe_webhook_secret: str = ""

    # App
    api_port: int = 8000
    rate_limit_per_minute: int = 30

    # Subscription quota (daily requests)
    quota_free: int = 10
    quota_pro: int = 200
    quota_premium: int = 1000


settings = Settings()
