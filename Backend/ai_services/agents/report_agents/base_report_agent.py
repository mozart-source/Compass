"""
Base class for report generation agents.
"""
from typing import Dict, Any, Optional, List, Union
import logging
from datetime import datetime, timedelta
import json

from ai_services.agents.base_agent import BaseAgent
from data_layer.models.report import Report, ReportStatus, ReportType
from data_layer.repos.report_repo import ReportRepository
from app.schemas.report_schemas import ReportProgressUpdate

logger = logging.getLogger(__name__)


class BaseReportAgent(BaseAgent):
    """
    Base agent for generating reports.

    This class extends the BaseAgent class and adds functionality
    specific to report generation, including:
    - Context gathering
    - Report formatting
    - Progress tracking
    - Error handling
    """

    report_type: ReportType = ReportType.SUMMARY  # Default, should be overridden
    name: str = "base_report_agent"
    description: str = "Base agent for generating reports"

    def __init__(self):
        """Initialize the base report agent."""
        super().__init__()
        self.report_repo = ReportRepository()

    async def gather_context(
        self,
        user_id: str,
        parameters: Dict[str, Any],
        time_range: Dict[str, str],
        auth_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Gather context data needed for report generation.

        Parameters:
            user_id (str): User ID to gather data for
            parameters (Dict[str, Any]): Additional parameters for gathering context
            time_range (Dict[str, str]): Time range for data (start_date, end_date)
            auth_token (Optional[str]): Authentication token

        Returns:
            Dict[str, Any]: Context data for report generation
        """
        # Base implementation - should be extended by subclasses
        return {
            "user_id": user_id,
            "parameters": parameters,
            "time_range": time_range,
            "timestamp": datetime.utcnow().isoformat()
        }

    async def format_report(
        self,
        raw_content: Union[str, Dict[str, Any]],
        context: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Format raw LLM output into structured report content.

        Parameters:
            raw_content (Union[str, Dict[str, Any]]): Raw content from LLM
            context (Dict[str, Any]): Context data used for generation

        Returns:
            Dict[str, Any]: Formatted report content
        """
        # If raw_content is already a dictionary, return it
        if isinstance(raw_content, dict):
            return raw_content

        # If raw_content is a string, try to parse it as JSON
        if isinstance(raw_content, str):
            try:
                # Check if the string looks like JSON
                if raw_content.strip().startswith('{') and raw_content.strip().endswith('}'):
                    return json.loads(raw_content)
                else:
                    # Not JSON, return as simple content
                    return {"content": raw_content}
            except json.JSONDecodeError as e:
                logger.error(f"Error parsing JSON content: {str(e)}")
                # Return as simple content if JSON parsing fails
                return {"content": raw_content}

        # Fallback for any other type
        return {"content": str(raw_content)}

    async def generate_report(
        self,
        report_id: str,
        user_id: str,
        parameters: Dict[str, Any],
        time_range: Dict[str, str],
        auth_token: Optional[str] = None,
        websocket: Optional[Any] = None
    ) -> Dict[str, Any]:
        """
        Generate a report based on provided parameters.

        Parameters:
            report_id (str): ID of the report to generate
            user_id (str): ID of the user to generate report for
            parameters (Dict[str, Any]): Parameters for report generation
            time_range (Dict[str, str]): Time range for report data
            auth_token (Optional[str]): Authentication token
            websocket (Optional[Any]): WebSocket connection for progress updates

        Returns:
            Dict[str, Any]: Generated report content
        """
        logger.info(
            f"Generating {self.report_type.value} report for user {user_id}")

        try:
            # Update report status to generating
            await self.report_repo.update_report_status(
                report_id,
                ReportStatus.GENERATING
            )

            # Send initial progress update
            if websocket:
                await self._send_progress_update(
                    websocket,
                    report_id,
                    0.1,
                    "Gathering context data...",
                    "generating"
                )

            # Gather context data
            context = await self.gather_context(
                user_id,
                parameters,
                time_range,
                auth_token
            )

            if websocket:
                await self._send_progress_update(
                    websocket,
                    report_id,
                    0.3,
                    "Analyzing data...",
                    "generating"
                )

            # Prepare prompt for LLM
            prompt = await self._prepare_report_prompt(context)

            if websocket:
                await self._send_progress_update(
                    websocket,
                    report_id,
                    0.5,
                    "Generating insights...",
                    "generating"
                )

            # Generate content with LLM
            raw_content = await self._generate_response(
                prompt,
                user_id,
                {"temperature": 0.2, "max_tokens": 2000}
            )

            if websocket:
                await self._send_progress_update(
                    websocket,
                    report_id,
                    0.8,
                    "Formatting report...",
                    "generating"
                )

            # Format the response into structured report content
            content = await self.format_report(raw_content, context)

            if websocket:
                await self._send_progress_update(
                    websocket,
                    report_id,
                    1.0,
                    "Report complete!",
                    "completed"
                )

            logger.info(f"Successfully generated report {report_id}")
            return content

        except Exception as e:
            error_msg = f"Error generating report: {str(e)}"
            logger.error(error_msg)

            # Update report status to failed
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
                    "Report generation failed",
                    "failed"
                )

            return {"error": error_msg}

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for report generation based on context.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        # Base implementation with a generic prompt
        # Should be overridden by subclasses for specific report types
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        prompt = f"""
        Generate a detailed {self.report_type.value} report for the user based on the following context:
        
        Time period: {start_date} to {end_date}
        
        The report should include:
        - A comprehensive summary of the user's data
        - Key insights and patterns
        - Actionable recommendations
        
        Format the report with clear sections including headings.
        
        Context data:
        {context}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief summary of key findings",
            "content": {{
                // Structured report content
            }},
            "sections": [
                {{
                    "title": "Section Title",
                    "content": "Section content...",
                    "type": "text"
                }}
                // Additional sections...
            ]
        }}
        """

        return prompt

    def _extract_summary(self, content: Union[str, Dict[str, Any]]) -> str:
        """Extract a summary from the raw content if not explicitly provided."""
        if isinstance(content, dict) and "summary" in content:
            return content["summary"]

        if isinstance(content, str):
            # Extract first paragraph or up to 200 characters
            summary = content.split("\n\n")[0]
            if len(summary) > 200:
                summary = summary[:197] + "..."
            return summary

        return "Report generated successfully"

    def _create_default_sections(self, content: Union[str, Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Create default sections if not explicitly provided."""
        if isinstance(content, dict) and "sections" in content:
            return content["sections"]

        if isinstance(content, str):
            # Try to split by markdown headings
            sections = []
            current_section = None
            current_content = []

            for line in content.split("\n"):
                if line.startswith("#"):
                    # Save previous section if exists
                    if current_section:
                        sections.append({
                            "title": current_section,
                            "content": "\n".join(current_content),
                            "type": "text"
                        })
                    # Start new section
                    current_section = line.replace("#", "").strip()
                    current_content = []
                else:
                    current_content.append(line)

            # Add last section
            if current_section:
                sections.append({
                    "title": current_section,
                    "content": "\n".join(current_content),
                    "type": "text"
                })

            # If no sections were found, create a default one
            if not sections:
                sections = [{
                    "title": "Report Content",
                    "content": content,
                    "type": "text"
                }]

            return sections

        # Default section if content is neither string nor dict with sections
        return [{
            "title": "Report Content",
            "content": "No detailed content available",
            "type": "text"
        }]

    async def _send_progress_update(
        self,
        websocket: Any,
        report_id: str,
        progress: float,
        message: str,
        status: str
    ) -> None:
        """
        Send a progress update through the WebSocket.

        Parameters:
            websocket (Any): WebSocket connection
            report_id (str): ID of the report being generated
            progress (float): Progress percentage (0.0-1.0)
            message (str): Status message
            status (str): Current report status
        """
        update = ReportProgressUpdate(
            report_id=report_id,
            progress=progress,
            status=status,
            message=message
        )

        await websocket.send_json(update.dict())
