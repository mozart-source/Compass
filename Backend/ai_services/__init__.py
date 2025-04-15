"""
AI services package for COMPASS.
"""

import logging
from ai_services.report_service import ReportService

logger = logging.getLogger(__name__)


def initialize_report_service():
    """Initialize the report service with available agents."""
    report_service = ReportService()
    logger.info("Report service initialized with available agents")
    return report_service


# Initialize report service
report_service = initialize_report_service()
