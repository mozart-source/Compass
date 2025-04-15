from typing import Optional, Any, Dict
from pymongo import MongoClient
from pymongo.database import Database
from pymongo.collection import Collection
from pymongo.errors import ConnectionFailure, ServerSelectionTimeoutError, OperationFailure
import logging
from motor.motor_asyncio import AsyncIOMotorClient
from functools import lru_cache
from core.config import settings
import ssl
import certifi

logger = logging.getLogger(__name__)

# Global client instances
_mongodb_client: Optional[MongoClient] = None
_async_mongodb_client: Optional[AsyncIOMotorClient] = None

# Connection pool settings from core.config
MAX_POOL_SIZE = settings.mongodb_max_pool_size
MIN_POOL_SIZE = settings.mongodb_min_pool_size
MAX_IDLE_TIME_MS = settings.mongodb_max_idle_time_ms
CONNECT_TIMEOUT_MS = settings.mongodb_connect_timeout_ms
SERVER_SELECTION_TIMEOUT_MS = settings.mongodb_server_selection_timeout_ms
MAX_CONNECTING = settings.mongodb_max_connecting
WAIT_QUEUE_TIMEOUT_MS = settings.mongodb_wait_queue_timeout_ms


@lru_cache(maxsize=1)
def get_mongodb_uri() -> str:
    """Get MongoDB URI from settings or environment variables."""
    return settings.mongodb_uri


def get_ssl_config() -> Dict[str, Any]:
    """Get SSL configuration for MongoDB connection."""
    # Basic configuration with increased timeouts for Docker
    ssl_config = {
        'retryWrites': True,
        'w': 'majority',
        'connectTimeoutMS': 15000,  # Increased timeout for Docker
        'serverSelectionTimeoutMS': 15000,  # Increased timeout for Docker
        'socketTimeoutMS': 15000,  # Increased timeout for Docker
        'maxPoolSize': 50,
        'minPoolSize': 5,
    }

    # For Docker environments, disable SSL entirely to fix compatibility issues
    if settings.DOCKER_ENV:
        logger.info(
            "Docker environment detected, using minimal config without SSL")
        ssl_config.update({
            'connectTimeoutMS': 30000,  # Even longer timeout for Docker
            'serverSelectionTimeoutMS': 30000,
        })
    else:
        # Production SSL settings
        ssl_config.update({
            'tls': True,
            'tlsCAFile': certifi.where(),
            'tlsAllowInvalidCertificates': False,
            'tlsAllowInvalidHostnames': False,
        })

    return ssl_config


def get_mongodb_client() -> Optional[MongoClient]:
    """Get the global MongoDB client instance with enhanced error handling."""
    global _mongodb_client

    if _mongodb_client is None:
        try:
            logger.info("Initializing MongoDB client...")

            # Get SSL configuration
            ssl_config = get_ssl_config()

            # Create client with enhanced configuration
            _mongodb_client = MongoClient(
                settings.mongodb_uri,
                **ssl_config
            )

            # Test the connection with retries
            max_retries = 3
            for attempt in range(max_retries):
                try:
                    # Test connection
                    _mongodb_client.admin.command('ping')
                    logger.info(
                        f"✅ MongoDB connection successful on attempt {attempt + 1}")
                    break
                except ServerSelectionTimeoutError as e:
                    if attempt == max_retries - 1:
                        logger.error(
                            f"❌ MongoDB connection failed after {max_retries} attempts: {e}")
                        # Don't raise, return None to handle gracefully
                        _mongodb_client = None
                        return None
                    else:
                        logger.warning(
                            f"⚠️ MongoDB connection attempt {attempt + 1} failed, retrying...")
                        continue
                except Exception as e:
                    logger.error(f"❌ Unexpected MongoDB connection error: {e}")
                    _mongodb_client = None
                    return None

        except Exception as e:
            logger.error(f"❌ Failed to initialize MongoDB client: {e}")
            _mongodb_client = None
            return None

    return _mongodb_client


def get_async_mongodb_client() -> Any:
    """Get async MongoDB client singleton with optimized connection pooling."""
    global _async_mongodb_client

    if _async_mongodb_client is None:
        uri = get_mongodb_uri()
        logger.info(f"Connecting to MongoDB (async) at {uri.split('@')[-1]}")

        try:
            # Configure async client with enhanced connection pooling
            _async_mongodb_client = AsyncIOMotorClient(
                uri,
                maxPoolSize=MAX_POOL_SIZE,
                minPoolSize=MIN_POOL_SIZE,
                maxIdleTimeMS=MAX_IDLE_TIME_MS,
                connectTimeoutMS=CONNECT_TIMEOUT_MS,
                serverSelectionTimeoutMS=SERVER_SELECTION_TIMEOUT_MS,
                retryWrites=True,
                w='majority',
                maxConnecting=MAX_CONNECTING,
                waitQueueTimeoutMS=WAIT_QUEUE_TIMEOUT_MS
            )
            logger.info(
                f"Successfully created async MongoDB client with pool size: {MAX_POOL_SIZE}")
        except Exception as e:
            logger.error(f"Failed to create async MongoDB client: {e}")
            raise

    return _async_mongodb_client


def get_database(db_name: Optional[str] = None) -> Optional[Database]:
    """Get MongoDB database."""
    client = get_mongodb_client()
    if client is None:
        logger.error("MongoDB client is not available")
        return None
    database_name = db_name or getattr(settings, 'mongodb_database', 'compass')
    return client[database_name]


def get_async_database(db_name: Optional[str] = None) -> Any:
    """Get async MongoDB database."""
    client = get_async_mongodb_client()
    database_name = db_name or getattr(settings, 'mongodb_database', 'compass')
    return client[database_name]


def get_collection(collection_name: str, db_name: Optional[str] = None) -> Optional[Collection]:
    """Get MongoDB collection."""
    db = get_database(db_name)
    if db is None:
        logger.error("Database is not available")
        return None
    return db[collection_name]


def get_async_collection(collection_name: str, db_name: Optional[str] = None) -> Any:
    """Get async MongoDB collection."""
    db = get_async_database(db_name)
    return db[collection_name]


async def close_mongodb_connections():
    """Close all MongoDB connections."""
    global _mongodb_client, _async_mongodb_client

    logger.info("Closing MongoDB connections")

    if _mongodb_client is not None:
        _mongodb_client.close()
        _mongodb_client = None
        logger.info("Closed sync MongoDB client")

    if _async_mongodb_client is not None:
        _async_mongodb_client.close()
        _async_mongodb_client = None
        logger.info("Closed async MongoDB client")
