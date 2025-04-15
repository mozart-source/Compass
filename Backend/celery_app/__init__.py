from celery import Celery
from Backend.core.config import settings

# Create Celery instance
celery_app = Celery(
    settings.APP_NAME.lower(),
    backend=settings.CELERY_RESULT_BACKEND,
    broker=settings.CELERY_BROKER_URL
)

# Import task modules
celery_app.config_from_object('Backend.celery_app.config')

# Auto-discover tasks in the tasks directory
celery_app.autodiscover_tasks(['Backend.celery_app.tasks'])

# Export the app for easy imports
__all__ = ['celery_app']