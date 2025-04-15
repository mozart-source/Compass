"""
Report service for generating and managing reports.
"""
import logging
from typing import Dict, Any, Optional, List, Type
from fastapi import WebSocket
from datetime import datetime, timedelta

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from ai_services.agents.report_agents.activity_report_agent import ActivityReportAgent
from ai_services.agents.report_agents.productivity_report_agent import ProductivityReportAgent
from ai_services.agents.report_agents.habits_report_agent import HabitsReportAgent
from ai_services.agents.report_agents.task_report_agent import TaskReportAgent
from ai_services.agents.report_agents.summary_report_agent import SummaryReportAgent
from ai_services.agents.report_agents.dashboard_report_agent import DashboardReportAgent
from ai_services.report.report_orchestrator import ReportOrchestrator
from ai_services.report.data_fetcher import DataFetcherService
from data_layer.repos.report_repo import ReportRepository
from data_layer.models.report import Report, ReportStatus, ReportType
from app.schemas.report_schemas import ReportCreate, ReportUpdate
from data_layer.cache.redis_client import redis_client
import json

logger = logging.getLogger(__name__)


class ReportService:
    """
    Service for generating and managing reports.

    This service integrates all report components and provides
    a unified interface for report operations.
    """

    def __init__(self):
        """Initialize the report service."""
        self.report_repo = ReportRepository()
        self.data_fetcher = DataFetcherService()
        self.orchestrator = ReportOrchestrator(
            self.report_repo, self.data_fetcher)

        # Register report agents
        self._register_agents()

    def _register_agents(self) -> None:
        """Register all report agents with the orchestrator."""
        self.orchestrator.register_agent(
            ReportType.ACTIVITY.value, ActivityReportAgent)
        self.orchestrator.register_agent(
            ReportType.PRODUCTIVITY.value, ProductivityReportAgent)
        self.orchestrator.register_agent(
            ReportType.HABITS.value, HabitsReportAgent)
        self.orchestrator.register_agent(
            ReportType.TASK.value, TaskReportAgent)
        self.orchestrator.register_agent(
            ReportType.SUMMARY.value, SummaryReportAgent)
        self.orchestrator.register_agent(
            ReportType.DASHBOARD.value, DashboardReportAgent)

        logger.info("Registered all report agents")

    async def create_report(
        self,
        user_id: str,
        report_data: ReportCreate
    ) -> Report:
        """
        Create a new report. If a similar report (completed or in progress)
        exists from the last 10 minutes, it will be returned instead of
        creating a new one.

        Args:
            user_id: ID of the user creating the report
            report_data: Report creation data

        Returns:
            Created report
        """
        # Determine time_range if not provided, normalizing for cacheability
        time_range = report_data.time_range
        if not time_range:
            # Normalize end_date to the minute to make it cacheable
            end_date = datetime.utcnow().replace(second=0, microsecond=0)
            # Default to 30 days for habits reports, 7 days for all others
            if report_data.type == ReportType.HABITS:
                start_date = end_date - timedelta(days=30)
            else:
                start_date = end_date - timedelta(days=7)

            time_range = {
                "start_date": start_date.isoformat(),
                "end_date": end_date.isoformat()
            }

        # Check if a similar, recently created report already exists (completed or in progress)
        search_since = datetime.utcnow() - timedelta(minutes=10)
        existing_report = await self.report_repo.async_find_one({
            "user_id": user_id,
            "type": report_data.type,
            "status": {"$in": [
                ReportStatus.COMPLETED,
                ReportStatus.GENERATING,
                ReportStatus.PENDING
            ]},
            "parameters": report_data.parameters or {},
            "time_range": time_range,
            "created_at": {"$gte": search_since}
        })

        if existing_report:
            logger.info(
                f"Found recent similar report {existing_report.id} with status '{existing_report.status.value}', returning it to client.")
            return existing_report

        # Create report data dictionary
        report_dict = {
            "user_id": user_id,
            "title": report_data.title,
            "type": report_data.type,
            "parameters": report_data.parameters,
            "time_range": time_range,
            "custom_prompt": report_data.custom_prompt
        }

        # Create report in database
        report = await self.report_repo.create_report_from_data(report_dict)

        logger.info(f"Created report {report.id} for user {user_id}")

        return report

    async def get_report(self, report_id: str) -> Optional[Report]:
        """
        Get a report by ID, with caching.

        Args:
            report_id: ID of the report to get

        Returns:
            Report if found, None otherwise
        """
        cache_key = f"report:{report_id}"
        cached_report = await redis_client.get(cache_key)
        if cached_report:
            logger.info(f"CACHE HIT for report {report_id}")
            return Report.parse_raw(cached_report)

        logger.info(f"CACHE MISS for report {report_id}")
        report = await self.report_repo.get_report(report_id)
        if report:
            await redis_client.set(cache_key, report.json(), ex=60)
        return report

    async def list_user_reports(
        self,
        user_id: str,
        skip: int = 0,
        limit: int = 20,
        status: Optional[str] = None,
        report_type: Optional[str] = None
    ) -> List[Report]:
        """
        List reports for a user, with caching.

        Args:
            user_id: ID of the user
            skip: Number of reports to skip
            limit: Maximum number of reports to return
            status: Optional filter by report status
            report_type: Optional filter by report type

        Returns:
            List of reports
        """
        cache_key = f"user_reports:{user_id}:{skip}:{limit}:{status}:{report_type}"
        cached_reports = await redis_client.get(cache_key)
        if cached_reports:
            logger.info(f"CACHE HIT for user reports list for user {user_id}")
            report_dicts = json.loads(cached_reports)
            return [Report(**data) for data in report_dicts]

        logger.info(f"CACHE MISS for user reports list for user {user_id}")
        reports = await self.report_repo.list_user_reports(
            user_id,
            status=status,
            report_type=report_type,
            skip=skip,
            limit=limit
        )
        if reports:
            report_dicts = [report.dict() for report in reports]
            await redis_client.set(cache_key, json.dumps(report_dicts, default=str), ex=60)
        return reports

    async def update_report(
        self,
        report_id: str,
        update_data: ReportUpdate
    ) -> Optional[Report]:
        """
        Update a report and invalidate relevant caches.

        Args:
            report_id: ID of the report to update
            update_data: Report update data

        Returns:
            Updated report if found, None otherwise
        """
        # Convert update_data to dict and remove None values
        update_dict = {k: v for k, v in update_data.dict().items()
                       if v is not None}

        if not update_dict:
            # Nothing to update
            return await self.report_repo.get_report(report_id)

        updated_report = await self.report_repo.update_report(report_id, update_dict)

        if updated_report:
            # Invalidate caches
            await redis_client.delete(f"report:{report_id}")
            user_report_keys = await redis_client.keys(f"user_reports:{updated_report.user_id}:*")
            if user_report_keys:
                await redis_client.delete(*user_report_keys)

        return updated_report

    async def delete_report(self, report_id: str) -> bool:
        """
        Delete a report and invalidate relevant caches.

        Args:
            report_id: ID of the report to delete

        Returns:
            True if deleted, False otherwise
        """
        # Get user_id before deleting for cache invalidation
        report = await self.get_report(report_id)
        if not report:
            return False
        user_id = report.user_id

        success = await self.report_repo.delete_report(report_id)

        if success:
            # Invalidate caches
            await redis_client.delete(f"report:{report_id}")
            user_report_keys = await redis_client.keys(f"user_reports:{user_id}:*")
            if user_report_keys:
                await redis_client.delete(*user_report_keys)

        return success

    async def generate_report(
        self,
        report_id: str,
        auth_token: Optional[str] = None,
        websocket: Optional[WebSocket] = None
    ) -> Dict[str, Any]:
        """
        Generate a report.

        Args:
            report_id: ID of the report to generate
            auth_token: Optional auth token for API access
            websocket: Optional WebSocket for progress updates

        Returns:
            Dict containing the generated report content or error
        """
        # Update report status to generating
        await self.report_repo.update_report_status(
            report_id,
            ReportStatus.GENERATING
        )

        # Generate report using orchestrator
        result = await self.orchestrator.generate_report(
            report_id,
            auth_token,
            websocket
        )

        if "error" in result:
            # Report generation failed
            await self.report_repo.update_report_status(
                report_id,
                ReportStatus.FAILED,
                error=result["error"]
            )
        else:
            # Report generation succeeded
            await self.report_repo.update_report_content(
                report_id,
                content=result.get("content", {}),
                summary=result.get("summary", ""),
                sections=result.get("sections", [])
            )

        return result

    async def regenerate_report(
        self,
        report_id: str,
        auth_token: Optional[str] = None,
        websocket: Optional[WebSocket] = None
    ) -> Dict[str, Any]:
        """
        Regenerate an existing report.

        Args:
            report_id: ID of the report to regenerate
            auth_token: Optional auth token for API access
            websocket: Optional WebSocket for progress updates

        Returns:
            Dict containing the generated report content or error
        """
        # Reset report status to generating
        await self.report_repo.update_report_status(
            report_id,
            ReportStatus.GENERATING
        )

        # Generate report using orchestrator
        return await self.generate_report(report_id, auth_token, websocket)

    async def list_available_report_types(self) -> List[Dict[str, str]]:
        """
        Get a list of available report types.

        Returns:
            List of dicts containing report type and description
        """
        return await self.orchestrator.list_available_report_types()
