"""
API routes for report generation and management.
"""

import logging
from typing import Dict, Any, Optional, List
from fastapi import APIRouter, Depends, HTTPException, WebSocket, status
from fastapi.responses import JSONResponse
from app.schemas.report_schemas import (
    ReportCreate, ReportUpdate, ReportResponse, ReportListResponse
)
from ai_services import report_service
from data_layer.repos.report_repo import ReportRepository
from data_layer.models.report import Report
from utils.jwt import extract_user_id_from_token

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/reports", tags=["Reports"])
report_repo = ReportRepository()


async def get_report_for_user(
    report_id: str,
    user_id: str = Depends(extract_user_id_from_token)
) -> Report:
    """
    Dependency to get a report and verify ownership.
    """
    report = await report_service.get_report(report_id)

    if not report:
        raise HTTPException(
            status_code=404,
            detail=f"Report with ID {report_id} not found"
        )

    if report.user_id != user_id:
        raise HTTPException(
            status_code=403,
            detail="You don't have permission to access this report"
        )

    return report


@router.post("", response_model=Dict[str, str])
async def create_report(
    report_data: ReportCreate,
    user_id: str = Depends(extract_user_id_from_token)
) -> Dict[str, str]:
    """
    Create a new report.

    Args:
        report_data: Report creation data
        user_id: Current authenticated user ID

    Returns:
        Dict with report ID
    """
    try:
        # Create report using the report_data object directly
        report = await report_service.create_report(
            user_id=user_id,
            report_data=report_data
        )

        # Ensure report.id is a string
        return {"report_id": str(report.id)}

    except Exception as e:
        logger.error(f"Error creating report: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Error creating report: {str(e)}"
        )


@router.get("/{report_id}", response_model=ReportResponse)
async def get_report(report: Report = Depends(get_report_for_user)) -> Any:
    """
    Get a report by ID.
    """
    return report


@router.get("", response_model=ReportListResponse)
async def list_reports(
    page: int = 1,
    limit: int = 10,
    report_type: Optional[str] = None,
    status: Optional[str] = None,
    user_id: str = Depends(extract_user_id_from_token)
) -> Any:
    """
    List reports for the current user.

    Args:
        page: Page number
        limit: Number of reports per page
        report_type: Filter by report type
        status: Filter by status
        user_id: Current authenticated user ID

    Returns:
        List of reports with pagination info
    """
    try:
        skip = (page - 1) * limit
        reports = await report_service.list_user_reports(
            user_id=user_id,
            skip=skip,
            limit=limit,
            report_type=report_type,
            status=status
        )

        # Create filter for counting total reports
        filter_dict = {"user_id": user_id}
        if report_type:
            filter_dict["type"] = report_type
        if status:
            filter_dict["status"] = status

        # Count total reports for pagination
        total_reports = await report_repo.async_count(filter_dict)

        # Create response with pagination info
        return {
            "reports": reports,
            "total": total_reports,
            "page": page,
            "limit": limit
        }

    except Exception as e:
        logger.error(f"Error listing reports: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Error listing reports: {str(e)}"
        )


@router.delete("/{report_id}", response_model=Dict[str, bool])
async def delete_report(report: Report = Depends(get_report_for_user)) -> Dict[str, bool]:
    """
    Delete a report.
    """
    try:
        assert report.id is not None, "Report ID cannot be None for deletion"
        success = await report_service.delete_report(str(report.id))
        return {"success": success}

    except Exception as e:
        logger.error(f"Error deleting report: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Error deleting report: {str(e)}"
        )


@router.patch("/{report_id}", response_model=ReportResponse)
async def update_report(
    update_data: ReportUpdate,
    report: Report = Depends(get_report_for_user)
) -> Any:
    """
    Update a report.
    """
    try:
        assert report.id is not None, "Report ID cannot be None for update"
        updated_report = await report_service.update_report(str(report.id), update_data)
        return updated_report

    except Exception as e:
        logger.error(f"Error updating report: {str(e)}")
        raise HTTPException(
            status_code=500,
            detail=f"Error updating report: {str(e)}"
        )
