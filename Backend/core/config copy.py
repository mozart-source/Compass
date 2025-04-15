from pydantic_settings import BaseSettings
from typing import List, Optional, Dict, Any
import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv()


class Settings(BaseSettings):
    # API Settings
    api_host: str = "127.0.0.1"
    api_port: int = 8001
    go_backend_path: str = "../Backend_go/cmd/mcp/main.go"

    # External Service URLs
    # In development, use localhost URLs
    # In production with Docker, use service names from docker-compose
    GO_BACKEND_URL: str = os.getenv("GO_BACKEND_URL",
                                    "http://compass-api-service:8000" if os.getenv("ENVIRONMENT") == "production"
                                    else "http://localhost:8000")
    NOTES_SERVER_URL: str = os.getenv("NOTES_SERVER_URL",
                                      "http://notes-server:5000" if os.getenv("ENVIRONMENT") == "production"
                                      else "http://localhost:5000")

    # Log the configured service URLs
    def __init__(self, **data):
        super().__init__(**data)
        import logging
        logger = logging.getLogger(__name__)
        logger.info(f"Configured GO_BACKEND_URL: {self.GO_BACKEND_URL}")
        logger.info(f"Configured NOTES_SERVER_URL: {self.NOTES_SERVER_URL}")

    # MCP Settings
    mcp_enabled: bool = True
    mcp_log_level: str = "INFO"
    mcp_sampling_enabled: bool = True
    mcp_server_host: str = "127.0.0.1"
    mcp_server_port: int = 8001
    go_backend_url: str = "../Backend_go/cmd/mcp/main.go"
    go_backend_api_key: Optional[str] = os.getenv("GO_BACKEND_API_KEY", None)
    mcp_api_key: Optional[str] = None

    # HTTPS Settings
    use_https: bool = os.getenv("USE_HTTPS", "false").lower() == "false"
    https_cert_file: str = os.getenv("HTTPS_CERT_FILE", "../certs/server.crt")
    https_key_file: str = os.getenv("HTTPS_KEY_FILE", "../certs/server.key")

    # PostgreSQL Database Settings
    db_user: str = "elhadi"
    db_password: str = "test123"
    db_host: str = "localhost"
    db_port: str = "5432"
    db_name: str = "compass"
    database_url: str = f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"

    # MongoDB Settings
    mongodb_uri: str = "mongodb+srv://ahmedelhadi1777:fb5OpNipjvS65euk@cluster0.ojy4aft.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
    mongodb_database: str = "compass_ai"
    mongodb_min_pool_size: int = 10
    mongodb_max_pool_size: int = 50
    mongodb_max_idle_time_ms: int = 30000
    mongodb_connect_timeout_ms: int = 5000
    mongodb_server_selection_timeout_ms: int = 5000

    # App Settings
    testing: bool = False
    debug: bool = True
    environment: str = "development"
    app_name: str = "COMPASS"
    app_version: str = "1.0.0"

    # JWT Settings
    jwt_secret_key: str = os.getenv(
        "JWT_SECRET", "a82552a2c8133eddce94cc781f716cdcb911d065528783a8a75256aff6731886")
    jwt_algorithm: str = os.getenv("JWT_ALGORITHM", "HS256")
    jwt_issuer: str = os.getenv("JWT_ISSUER", "compass")
    access_token_expire_minutes: int = int(
        os.getenv("JWT_EXPIRY_HOURS", "24")) * 60  # Convert hours to minutes

    # Database Init Settings
    db_init_mode: str = "dev"
    create_admin: bool = True
    admin_email: str = "admin@aiwa.com"
    admin_username: str = "admin"
    admin_password: str = "aiwa_admin_2024!"
    db_drop_tables: bool = True

    # API Settings
    api_v1_prefix: str = "/api/v1"
    cors_origins: List[str] = ["http://localhost:3000", "http://localhost:8080"]

    # Logging Settings
    log_level: str = "INFO"
    log_format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"

    # Path Settings
    pythonpath: str = ""

    # Redis Settings
    redis_host: str = "localhost"  # for docker change to redis
    redis_port: int = 6380
    redis_db: int = 1
    redis_password: str = os.getenv("REDIS_PASSWORD", "")
    redis_url: str = f"redis://:{redis_password}@{redis_host}:{redis_port}/{redis_db}"

    # AI Service Settings
    ai_services_enabled: bool = True
    ai_model_path: str = "models"
    ai_cache_ttl: int = 3600
    ai_max_tokens: int = 2048
    ai_temperature: float = 0.7
    ai_top_p: float = 0.9

    # Atomic Agents Settings
    atomic_agents_enabled: bool = True
    openai_api_key: str = os.getenv(
        "OPENAI_API_KEY", "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")
    atomic_agents_default_model: str = "gpt-4o-mini"
    atomic_agents_streaming: bool = True
    atomic_agents_memory_enabled: bool = True
    atomic_agents_memory_max_tokens: int = 4000
    atomic_agents_memory_max_messages: int = 10
    atomic_agents_providers: Dict[str, str] = {
        "github": "gpt-4o-mini",
        "anthropic": "claude-3-haiku-20240307"
    }

    # LLM Settings
    llm_api_key: str = "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"
    llm_api_base_url: str = "https://models.github.ai/inference"
    llm_model_name: str = "openai/gpt-4.1-mini"
    llm_temperature: float = 0.7
    llm_max_tokens: int = 4096
    llm_top_p: float = 0.95
    llm_min_p: float = 0.0
    llm_top_k: int = 0

    # LLM Billing Settings
    # pay-as-you-go, provisioned, quota-based, hybrid
    llm_billing_type: str = "pay-as-you-go"
    llm_token_cost_input: float = 0.25  # Cost per million input tokens
    llm_token_cost_output: float = 0.75  # Cost per million output tokens
    llm_pricing_model: str = "per_token"  # per_token, per_request, per_hour
    # Tokens per hour for provisioned capacity
    llm_provisioned_capacity: Optional[int] = None
    # Cost per hour for provisioned capacity
    llm_provisioned_cost: Optional[float] = None
    llm_quota_limit: Optional[int] = None  # Token quota limit
    llm_quota_reset_interval: str = "monthly"  # daily, weekly, monthly
    llm_enable_quotas: bool = True  # Whether to enable quota tracking
    llm_max_context_tokens: int = 16000  # Maximum context window size
    llm_max_output_tokens: int = 4096  # Maximum output tokens per request

    # Model Pricing Configuration
    model_pricing: Dict[str, Dict[str, Any]] = {
        "openai/gpt-4.1": {
            "input_cost_per_million": 0.25,
            "output_cost_per_million": 0.75,
            "billing_type": "pay-as-you-go",
            "max_context_tokens": 16000,
            "max_output_tokens": 4096
        },
        "claude-3-sonnet-20240229": {
            "input_cost_per_million": 0.15,
            "output_cost_per_million": 0.50,
            "billing_type": "pay-as-you-go",
            "max_context_tokens": 200000,
            "max_output_tokens": 4096
        }
    }

    # Billing Quota Settings
    billing_quota_enabled: bool = True
    billing_quota_default_limit: int = 1000000
    billing_quota_reset_interval: str = "monthly"
    billing_quota_grace_period: str = "24h"
    billing_quota_alert_threshold: float = 0.8
    billing_quota_alert_emails: List[str] = ["admin@example.com"]

    # Cost Tracking Settings
    cost_tracking_enabled: bool = True
    cost_tracking_interval: str = "hourly"
    cost_tracking_retention_days: int = 90
    cost_tracking_alert_threshold: float = 100.0
    cost_tracking_alert_emails: List[str] = ["admin@example.com"]

    # Dashboard Real-time Update Settings (Optimized for Instant UX)
    dashboard_realtime_enabled: bool = os.getenv(
        "ENABLE_REALTIME_UPDATES", "true").lower() == "true"
    dashboard_update_throttle_seconds: float = float(
        os.getenv("UPDATE_THROTTLE_SECONDS", "0.05"))  # 50ms for instant feel
    dashboard_batch_updates: bool = os.getenv(
        "BATCH_UPDATES", "false").lower() == "true"  # Disabled for instant response
    dashboard_quiet_mode: bool = os.getenv(
        "DASHBOARD_QUIET_MODE", "false").lower() == "true"
    dashboard_dedup_window: float = float(
        os.getenv("DASHBOARD_DEDUP_WINDOW", "0.3"))  # 300ms deduplication

    # WebSocket Settings for Dashboard
    websocket_ping_interval: int = int(
        os.getenv("WEBSOCKET_PING_INTERVAL", "15"))  # 15 seconds
    websocket_connection_timeout: int = int(
        os.getenv("WEBSOCKET_CONNECTION_TIMEOUT", "3"))  # 3 seconds
    websocket_max_reconnection_delay: int = int(
        os.getenv("WEBSOCKET_MAX_RECONNECTION_DELAY", "5"))  # 5 seconds

    # Usage Analytics Settings
    usage_analytics_enabled: bool = True
    usage_analytics_interval: str = "hourly"
    usage_analytics_retention_days: int = 90
    usage_analytics_export_enabled: bool = True
    usage_analytics_export_format: str = "csv"
    usage_analytics_export_path: str = "data/analytics"

    # NLP Settings
    nlp_model_name: str = "flan-t5-base"
    nlp_api_key: str = "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"
    nlp_api_base_url: str = "URL_ADDRESS.deepinfra.com/v1/openai"
    nlp_batch_size: int = 32

    # Emotion Analysis Settings
    emotion_model_name: str = "j-hartmann/emotion-english-distilroberta-base"
    emotion_api_key: str = "your_huggingface_api_key"
    emotion_api_base_url: str = "URL_ADDRESS-inference.huggingface.co"
    emotion_threshold: float = 0.5

    # ChromaDB Settings
    chroma_collection_name: str = "compass_knowledge_base"
    chroma_db_path: str = "./data/chromadb"

    # Embedding Settings
    embedding_model_name: str = "all-MiniLM-L6-v2"
    embedding_api_key: str = "your_huggingface_api_key"

    # GitHub Model Adapter Settings
    github_adapter_enabled: bool = True
    github_adapter_debug: bool = True
    github_adapter_timeout: int = 60
    github_adapter_model_name: str = "openai/gpt-4.1-mini"
    github_adapter_fallback: bool = True

    def log_dashboard_config(self):
        """Log dashboard configuration for debugging"""
        import logging
        logger = logging.getLogger(__name__)
        logger.info(f"Dashboard Configuration:")
        logger.info(f"  Real-time updates: {self.dashboard_realtime_enabled}")
        logger.info(
            f"  Update throttle: {self.dashboard_update_throttle_seconds}s")
        logger.info(f"  Batch updates: {self.dashboard_batch_updates}")
        logger.info(f"  Quiet mode: {self.dashboard_quiet_mode}")
        logger.info(f"  Deduplication window: {self.dashboard_dedup_window}s")
        logger.info(
            f"  WebSocket ping interval: {self.websocket_ping_interval}s")

    model_config = {
        "env_file": ".env",
        "case_sensitive": True,
        "extra": "allow",
        "env_prefix": "",
        "use_enum_values": True,
    }


settings = Settings()
