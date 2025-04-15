from Backend.celery_app import celery_app
from typing import List, Dict, Optional
from datetime import datetime


@celery_app.task(
    name="tasks.email_tasks.send_email",
    queue="email",
    priority=5,
    retry_backoff=True,
    max_retries=3
)
def send_email(
    to_email: str,
    subject: str,
    body: str,
    attachments: Optional[List[Dict]] = None
) -> Dict:
    """
    Send an email asynchronously using Celery.
    """
    try:
        # TODO: Implement actual email sending logic
        return {
            "status": "success",
            "message": f"Email sent to {to_email}",
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        # Retry the task with exponential backoff
        send_email.retry(exc=e)


@celery_app.task(
    name="tasks.email_tasks.send_bulk_emails",
    queue="email",
    priority=7
)
def send_bulk_emails(emails: List[Dict]) -> List[Dict]:
    """
    Send multiple emails in bulk.
    """
    results = []
    for email in emails:
        try:
            result = send_email.delay(
                to_email=email["to"],
                subject=email["subject"],
                body=email["body"],
                attachments=email.get("attachments")
            )
            results.append({
                "email": email["to"],
                "task_id": result.id,
                "status": "queued"
            })
        except Exception as e:
            results.append({
                "email": email["to"],
                "error": str(e),
                "status": "failed"
            })
    return results
