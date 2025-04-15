"""Module for storing global MCP client state."""
from typing import Optional, Any
from mcp_py.client import MCPClient

# Global MCP client instance
mcp_client: Optional[MCPClient] = None
# Global MCP server process
mcp_process: Optional[Any] = None


def get_mcp_client() -> Optional[MCPClient]:
    """Get the global MCP client instance."""
    return mcp_client


def set_mcp_client(client: Optional[MCPClient]) -> None:
    """Set the global MCP client instance."""
    global mcp_client
    mcp_client = client


def get_mcp_process() -> Optional[Any]:
    """Get the global MCP server process."""
    return mcp_process


def set_mcp_process(process: Any) -> None:
    """Set the global MCP server process."""
    global mcp_process
    mcp_process = process
