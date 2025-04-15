import logging
import asyncio
from contextlib import asynccontextmanager
from fastapi import FastAPI
from data_layer.mongodb.connection import (
    get_mongodb_client,
    get_async_mongodb_client,
    close_mongodb_connections,
    get_database
)
from core.config import settings

logger = logging.getLogger(__name__)


async def init_collections():
    """Initialize MongoDB collections on startup."""
    try:
        logger.info("Initializing MongoDB collections...")
        db = get_database(settings.mongodb_database)

        # Define collections to create if they don't exist
        collections = [
            "ai_models",
            "model_usages",
            "conversations",
            "focus_sessions",
            "goals",
            "system_metrics",
            # future collections here
        ]

        # Get existing collections
        existing_collections = db.list_collection_names()
        logger.info(f"Existing collections: {existing_collections}")

        # Create collections that don't exist
        for collection in collections:
            if collection not in existing_collections:
                logger.info(f"Creating collection: {collection}")
                # Create with a dummy document (will be removed after)
                db[collection].insert_one({"_init": True})
                db[collection].delete_one({"_init": True})
                logger.info(f"Collection {collection} created")
            else:
                logger.info(f"Collection {collection} already exists")

        logger.info("MongoDB collections initialized successfully")
    except Exception as e:
        logger.error(f"Error initializing MongoDB collections: {str(e)}")
        # Don't raise - we want the application to start anyway


@asynccontextmanager
async def mongodb_lifespan(app: FastAPI):
    """Context manager for MongoDB lifecycle."""
    try:
        logger.info("Connecting to MongoDB...")

        # Initialize clients
        get_mongodb_client()
        get_async_mongodb_client()

        # Initialize collections
        await init_collections()

        logger.info("MongoDB connection initialized")
        yield
    finally:
        logger.info("Closing MongoDB connections...")
        await close_mongodb_connections()
        logger.info("MongoDB connections closed")
