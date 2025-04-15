from typing import Dict, Any, Optional, List, Union
from data_layer.repos.base_repo import BaseMongoRepository
from data_layer.models.report import Report, ReportStatus, ReportType
from datetime import datetime
import logging
import json

logger = logging.getLogger(__name__)


class ReportRepository(BaseMongoRepository[Report]):
    """Repository for managing reports in MongoDB."""

    def __init__(self):
        """Initialize the repository with the Report model."""
        super().__init__(Report)

    async def async_find_one(self, filter: Dict[str, Any]) -> Optional[Report]:
        """Find a single document by a filter dictionary."""
        document = await self.async_collection.find_one(filter)
        if document:
            return self.model(**document)
        return None

    async def create_report(
        self,
        user_id: str,
        title: str,
        report_type: ReportType,
        parameters: Optional[Dict[str, Any]] = None,
        time_range: Optional[Dict[str, str]] = None,
        custom_prompt: Optional[str] = None
    ) -> Report:
        """
        Create a new report.

        Args:
            user_id: User ID
            title: Report title
            report_type: Type of report
            parameters: Optional parameters for report generation
            time_range: Optional time range for report data
            custom_prompt: Optional custom prompt for report generation

        Returns:
            Created report
        """
        # Create report data dictionary
        report_data = {
            "title": title,
            "user_id": user_id,
            "type": report_type
        }

        # Add optional fields if provided
        if parameters:
            report_data["parameters"] = parameters

        if time_range:
            report_data["time_range"] = time_range

        if custom_prompt:
            report_data["custom_prompt"] = custom_prompt

        # Create the report instance
        report = Report(**report_data)

        # Insert the report
        report_id = await self.async_insert(report)

        # Return the created report
        created_report = await self.async_find_by_id(report_id)
        if created_report is None:
            # This should never happen, but to satisfy the linter
            raise ValueError(
                f"Failed to retrieve created report with ID {report_id}")
        return created_report

    async def create_report_from_data(self, report_data: Dict[str, Any]) -> Report:
        """
        Create a new report from a data dictionary.

        Args:
            report_data: Dictionary containing report data

        Returns:
            Created report
        """
        return await self.create_report(
            user_id=report_data["user_id"],
            title=report_data["title"],
            report_type=report_data["type"],
            parameters=report_data.get("parameters"),
            time_range=report_data.get("time_range"),
            custom_prompt=report_data.get("custom_prompt")
        )

    async def get_report(self, report_id: str) -> Optional[Report]:
        """Get a report by ID."""
        return await self.async_find_by_id(report_id)

    async def list_user_reports(
        self,
        user_id: str,
        status: Optional[str] = None,
        report_type: Optional[str] = None,
        skip: int = 0,
        limit: int = 20,
    ) -> List[Report]:
        """
        List reports for a user with optional filtering.

        Args:
            user_id: User ID
            status: Optional report status filter
            report_type: Optional report type filter
            skip: Number of reports to skip
            limit: Maximum number of reports to return

        Returns:
            List of reports
        """
        filter_dict: Dict[str, Any] = {"user_id": user_id}

        if status:
            try:
                filter_dict["status"] = ReportStatus(status)
            except ValueError:
                # If invalid status, ignore the filter
                pass

        if report_type:
            try:
                filter_dict["type"] = ReportType(report_type)
            except ValueError:
                # If invalid type, ignore the filter
                pass

        return await self.async_find_many(filter_dict, skip=skip, limit=limit)

    async def update_report_status(
        self,
        report_id: str,
        status: ReportStatus,
        error: Optional[str] = None
    ) -> Optional[Report]:
        """Update the status of a report."""
        # Create update data dictionary
        update_data: Dict[str, Any] = {"status": status}

        # Add completed_at timestamp if status is COMPLETED
        if status == ReportStatus.COMPLETED:
            update_data["completed_at"] = datetime.utcnow()

        # Add error message if provided
        if error:
            update_data["error"] = error

        # Update the report
        return await self.async_update(report_id, update_data)

    async def update_report_content(
        self,
        report_id: str,
        content: Union[Dict[str, Any], str],
        summary: Optional[str] = None,
        sections: Optional[List[Dict[str, Any]]] = None
    ) -> Optional[Report]:
        """Update the content of a report."""
        # Ensure content is a dictionary
        content_dict: Dict[str, Any] = {}

        if isinstance(content, dict):
            content_dict = content
        elif isinstance(content, str):
            # Try to parse string content as JSON
            try:
                if content.strip().startswith('{') and content.strip().endswith('}'):
                    content_dict = json.loads(content)
                else:
                    # Not JSON, store as simple content
                    content_dict = {"text": content}
            except json.JSONDecodeError as e:
                logger.error(f"Error parsing JSON content: {str(e)}")
                # Store as simple content if JSON parsing fails
                content_dict = {"text": content, "parse_error": str(e)}
        else:
            # Convert any other type to string and store
            content_dict = {"text": str(content)}

        # Create update data dictionary
        update_data: Dict[str, Any] = {
            "content": content_dict,
            "status": ReportStatus.COMPLETED,
            "completed_at": datetime.utcnow()
        }

        # Add summary if provided
        if summary:
            update_data["summary"] = summary

        # Add sections if provided
        if sections:
            update_data["sections"] = sections

        # Update the report
        return await self.async_update(report_id, update_data)

    async def get_reports(
        self,
        filters: Dict[str, Any],
        skip: int = 0,
        limit: int = 10
    ) -> Dict[str, Any]:
        """Get reports with pagination and filtering."""
        # Get reports matching filters
        reports = await self.async_find_many(filters, skip=skip, limit=limit)

        # Count total reports matching filters
        total = await self.async_count(filters)

        # Return reports and pagination info
        return {
            "reports": reports,
            "total": total
        }

    async def update_report(
        self,
        report_id: str,
        update_data: Dict[str, Any]
    ) -> Optional[Report]:
        """Update report data."""
        return await self.async_update(report_id, update_data)

    async def delete_report(self, report_id: str) -> bool:
        """Delete a report."""
        return await self.async_delete(report_id)

    async def delete_user_reports(self, user_id: str) -> int:
        """Delete all reports for a user."""
        return await self.async_delete_many({"user_id": user_id})
