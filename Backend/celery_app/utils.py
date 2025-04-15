import asyncio
from typing import AsyncGenerator, Callable, Any
from functools import wraps
import logging

logger = logging.getLogger(__name__)

def get_or_create_eventloop():
    """Get the current event loop or create a new one if none exists."""
    try:
        return asyncio.get_event_loop()
    except RuntimeError as ex:
        if "There is no current event loop in thread" in str(ex):
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            return loop
        raise

def async_to_sync(async_func: Callable) -> Callable:
    """Decorator to convert an async function to a sync function."""
    @wraps(async_func)
    def sync_wrapper(*args, **kwargs):
        loop = get_or_create_eventloop()
        return loop.run_until_complete(async_func(*args, **kwargs))
    return sync_wrapper

def task_with_retry(max_retries=3, countdown=60):
    """Decorator to create a task with retry logic."""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                logger.error(f"Task error: {str(e)}")
                raise
        return wrapper
    return decorator