from celery.signals import task_success, task_failure, task_retry, task_revoked
import logging

logger = logging.getLogger(__name__)

@task_success.connect
def task_success_handler(sender=None, **kwargs):
    """Log successful task execution."""
    logger.info(f"Task {sender.name} completed successfully")

@task_failure.connect
def task_failure_handler(sender=None, task_id=None, exception=None, **kwargs):
    """Log failed task execution."""
    logger.error(f"Task {sender.name} (id: {task_id}) failed: {str(exception)}")

@task_retry.connect
def task_retry_handler(sender=None, request=None, reason=None, **kwargs):
    """Log task retry."""
    logger.warning(f"Task {sender.name} being retried: {reason}")

@task_revoked.connect
def task_revoked_handler(sender=None, request=None, terminated=None, **kwargs):
    """Log revoked task."""
    logger.warning(f"Task {sender.name} was revoked")

def get_task_status(task_id):
    """Get the status of a task by its ID."""
    from Backend.celery_app import celery_app
    result = celery_app.AsyncResult(task_id)
    return {
        "task_id": task_id,
        "status": result.status,
        "result": result.result if result.ready() else None
    }