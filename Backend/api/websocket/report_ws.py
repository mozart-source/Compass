"""
WebSocket handler for real-time report generation updates.
"""
import logging
import json
from typing import Dict, Any, Optional
from fastapi import APIRouter, WebSocket, WebSocketDisconnect, Depends, HTTPException, status

from ai_services.report_service import ReportService
from data_layer.repos.report_repo import ReportRepository
from core.auth.jwt_handler import get_token_from_websocket, get_current_user

logger = logging.getLogger(__name__)

router = APIRouter()
report_service = ReportService()
report_repo = ReportRepository()


@router.websocket("/reports/{report_id}")
@router.websocket("/reports/{report_id}/generate")
async def report_websocket(websocket: WebSocket, report_id: str):
    """
    WebSocket endpoint for real-time report generation updates.

    Args:
        websocket: WebSocket connection
        report_id: ID of the report to generate
    """
    await websocket.accept()

    try:
        # Authenticate WebSocket connection and retrieve user/report details
        token_data = await get_token_from_websocket(websocket)
        if not token_data:
            return

        user = get_current_user(token_data)
        user_id = user["user_id"]
        token = token_data["raw"]

        report = await report_repo.get_report(report_id)
        if not report:
            logger.error(f"Report with ID {report_id} not found")
            await websocket.send_json({
                "status": "error",
                "message": f"Report with ID {report_id} not found"
            })
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
            return

        if report.user_id != user_id:
            logger.error(
                f"User {user_id} does not have access to report {report_id}")
            await websocket.send_json({
                "status": "error",
                "message": "You do not have access to this report"
            })
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
            return

    except Exception as auth_error:
        logger.error(
            f"WebSocket authentication or authorization failed: {auth_error}", exc_info=True)
        try:
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
        except:
            pass
        return

    try:
        # Send initial status update
        await websocket.send_json({
            "status": "connected",
            "report_id": report_id,
            "message": "Connected to report generation stream"
        })

        # Wait for commands from client
        while True:
            data = await websocket.receive_text()
            command = json.loads(data)

            if command.get("action") == "generate":
                # Start report generation
                await websocket.send_json({
                    "status": "generating",
                    "report_id": report_id,
                    "progress": 0.0,
                    "message": "Starting report generation..."
                })

                # Generate report
                result = await report_service.generate_report(
                    report_id,
                    auth_token=token,
                    websocket=websocket
                )

                if "error" in result:
                    await websocket.send_json({
                        "status": "failed",
                        "report_id": report_id,
                        "progress": 1.0,
                        "message": f"Report generation failed: {result['error']}"
                    })
                else:
                    await websocket.send_json({
                        "status": "completed",
                        "report_id": report_id,
                        "progress": 1.0,
                        "message": "Report generation completed",
                        "report": {
                            "id": report_id,
                            "summary": result.get("summary", ""),
                            "content": result.get("content", {})
                        }
                    })

            elif command.get("action") == "cancel":
                # Cancel report generation (not implemented yet)
                await websocket.send_json({
                    "status": "cancelled",
                    "report_id": report_id,
                    "message": "Report generation cancelled"
                })
                break

            elif command.get("action") == "ping":
                # Keep-alive ping
                await websocket.send_json({
                    "status": "pong",
                    "report_id": report_id
                })

            else:
                # Unknown command
                await websocket.send_json({
                    "status": "error",
                    "message": f"Unknown command: {command.get('action')}"
                })

    except WebSocketDisconnect:
        logger.info(f"WebSocket disconnected for report {report_id}")

    except Exception as e:
        logger.exception(f"Error in report WebSocket: {str(e)}")

        try:
            await websocket.send_json({
                "status": "error",
                "message": f"Internal server error: {str(e)}"
            })
        except:
            pass
