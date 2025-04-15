#!/usr/bin/env python
from ai_services.agents.report_agents.task_report_agent import TaskReportAgent
from ai_services.agents.report_agents.summary_report_agent import SummaryReportAgent
from ai_services.agents.report_agents.productivity_report_agent import ProductivityReportAgent
from ai_services.agents.report_agents.habits_report_agent import HabitsReportAgent
from ai_services.agents.report_agents.activity_report_agent import ActivityReportAgent
from mcp_py.client import MCPClient
from core.mcp_state import get_mcp_client, set_mcp_client
import asyncio
import sys
import os
import logging
from pathlib import Path

# Add the parent directory to the Python path
sys.path.insert(0, str(Path(__file__).parent.parent))

# Configure logging
logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Import necessary modules


async def test_mcp_client_initialization():
    """Test that the MCP client is properly initialized and accessible to report agents."""
    logger.info("Testing MCP client initialization...")

    # Initialize MCP client
    server_script = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                                 "mcp_py", "server.py")
    logger.info(f"Using MCP server script at: {server_script}")

    client = MCPClient()
    try:
        # Connect to the server
        await client.connect_to_server(server_script)
        logger.info("Successfully connected to MCP server")

        # Store the client in global state
        set_mcp_client(client)
        logger.info("Set MCP client in global state")

        # Verify that the client is accessible
        global_client = get_mcp_client()
        if global_client is not None:
            logger.info("Successfully retrieved MCP client from global state")
        else:
            logger.error("Failed to retrieve MCP client from global state")
            return False

        # Test that the client can be accessed from report agents
        agents = [
            ActivityReportAgent(),
            HabitsReportAgent(),
            ProductivityReportAgent(),
            SummaryReportAgent(),
            TaskReportAgent()
        ]

        for agent in agents:
            agent_client = await agent._get_mcp_client()
            if agent_client is not None:
                logger.info(
                    f"Agent {agent.name} successfully accessed MCP client")
            else:
                logger.error(f"Agent {agent.name} failed to access MCP client")
                return False

        # Test that the client can call a tool
        try:
            result = await client.call_tool("check.health")
            logger.info(f"Successfully called check.health: {result}")
        except Exception as e:
            logger.error(f"Error calling tool: {str(e)}")
            return False

        logger.info("All MCP client tests passed!")
        return True

    except Exception as e:
        logger.error(f"Error in MCP client initialization: {str(e)}")
        return False
    finally:
        # Clean up
        if client:
            await client.cleanup()
            logger.info("Cleaned up MCP client")


if __name__ == "__main__":
    success = asyncio.run(test_mcp_client_initialization())
    sys.exit(0 if success else 1)
