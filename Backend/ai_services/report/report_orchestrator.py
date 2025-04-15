"""
Report orchestrator for managing the report generation workflow.
"""

import logging
import asyncio
from typing import Dict, Any, Optional, Type, List
from fastapi import WebSocket

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from ai_services.report.data_fetcher import DataFetcherService
from data_layer.repos.report_repo import ReportRepository
from data_layer.models.report import Report, ReportStatus, ReportType
from app.schemas.report_schemas import ReportProgressUpdate

logger = logging.getLogger(__name__)


class ReportOrchestrator:
    """
    Orchestrator for report generation workflow.

    This class coordinates the report generation process by:
    - Selecting the appropriate agent based on report type
    - Managing the generation workflow
    - Handling errors and retries
    - Sending progress updates via WebSocket
    """

    def __init__(
        self,
        report_repo: Optional[ReportRepository] = None,
        data_fetcher: Optional[DataFetcherService] = None
    ):
        """
        Initialize the report orchestrator.

        Args:
            report_repo: Optional report repository instance
            data_fetcher: Optional data fetcher service instance
        """
        self.report_repo = report_repo or ReportRepository()
        self.data_fetcher = data_fetcher or DataFetcherService()
        self.report_agents: Dict[str, Type[BaseReportAgent]] = {}
        self.max_retries = 2

    def register_agent(self, report_type: str, agent_cls: Type[BaseReportAgent]) -> None:
        """
        Register a report agent for a specific report type.

        Args:
            report_type: Type of report the agent handles
            agent_cls: Agent class to register
        """
        self.report_agents[report_type] = agent_cls
        logger.info(
            f"Registered agent {agent_cls.__name__} for report type {report_type}")

    async def generate_report(
        self,
        report_id: str,
        auth_token: Optional[str] = None,
        websocket: Optional[WebSocket] = None
    ) -> Dict[str, Any]:
        """
        Orchestrate the report generation process.

        Args:
            report_id: ID of the report to generate
            auth_token: Optional auth token for API access
            websocket: Optional WebSocket for progress updates

        Returns:
            Dict containing the generated report content or error
        """
        # Get report from database
        report = await self.report_repo.get_report(report_id)

        if not report:
            error = f"Report with ID {report_id} not found"
            logger.error(error)
            return {"error": error}

        # Get appropriate agent for report type
        report_type = report.type.value

        if report_type not in self.report_agents:
            error = f"No agent registered for report type {report_type}"
            logger.error(error)

            # Update report status to failed
            await self.report_repo.update_report_status(
                report_id,
                ReportStatus.FAILED,
                error=error
            )

            return {"error": error}

        # Initialize retry counter
        retries = 0

        while retries <= self.max_retries:
            try:
                # Send progress update if this is a retry
                if retries > 0 and websocket:
                    await self._send_progress_update(
                        websocket,
                        report_id,
                        0.1,
                        f"Retrying report generation (attempt {retries+1}/{self.max_retries+1})...",
                        "generating"
                    )

                # Instantiate agent
                agent_cls = self.report_agents[report_type]
                agent = agent_cls()

                # Generate report
                result = await agent.generate_report(
                    report_id=report_id,
                    user_id=report.user_id,
                    parameters=report.parameters,
                    time_range=report.time_range,
                    auth_token=auth_token,
                    websocket=websocket
                )

                return result

            except Exception as e:
                error_msg = f"Error generating report (attempt {retries+1}/{self.max_retries+1}): {str(e)}"
                logger.error(error_msg)

                retries += 1

                # If we've reached max retries, mark as failed
                if retries > self.max_retries:
                    await self.report_repo.update_report_status(
                        report_id,
                        ReportStatus.FAILED,
                        error=error_msg
                    )

                    if websocket:
                        await self._send_progress_update(
                            websocket,
                            report_id,
                            1.0,
                            "Report generation failed after multiple attempts",
                            "failed"
                        )

                    return {"error": error_msg}

                # Otherwise wait and retry
                await asyncio.sleep(2 ** retries)  # Exponential backoff

        # This line should never be reached due to the loop structure,
        return {"error": "Unexpected error in report generation process"}

    async def list_available_report_types(self) -> List[Dict[str, str]]:
        """
        Get a list of available report types with descriptions.

        Returns:
            List of dicts containing report type and description
        """
        result = []

        for report_type, agent_cls in self.report_agents.items():
            result.append({
                "type": report_type,
                "name": agent_cls.name,
                "description": agent_cls.description
            })

        return result

    async def _send_progress_update(
        self,
        websocket: WebSocket,
        report_id: str,
        progress: float,
        message: str,
        status: str
    ) -> None:
        """
        Send a progress update through the WebSocket.

        Args:
            websocket: WebSocket connection
            report_id: ID of the report being generated
            progress: Progress percentage (0.0-1.0)
            message: Status message
            status: Current report status
        """
        update = ReportProgressUpdate(
            report_id=report_id,
            progress=progress,
            status=status,
            message=message
        )

        await websocket.send_json(update.dict())
