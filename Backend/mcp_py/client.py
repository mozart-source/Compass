from typing import Optional, Dict, Any, List, Callable, Awaitable, Union, AsyncIterator
import logging
import json
import os
import asyncio
import sys
from contextlib import AsyncExitStack
from mcp.client.stdio import stdio_client
from mcp.client.session import ClientSession
from mcp import StdioServerParameters
from core.config import settings
import time

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('mcp_client.log')
    ]
)
logger = logging.getLogger(__name__)


class Tool:
    """Represents an MCP tool with its metadata."""

    def __init__(self, name: str, description: str, input_schema: Dict[str, Any]):
        """Initialize a Tool object.

        Args:
            name: The name of the tool
            description: The description of the tool
            input_schema: The input schema for the tool
        """
        self.name = name
        self.description = description
        self.input_schema = input_schema

    def to_dict(self) -> Dict[str, Any]:
        """Convert tool to dictionary representation."""
        return {
            "name": self.name,
            "description": self.description,
            "input_schema": self.input_schema
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Tool':
        """Create a Tool from dictionary data."""
        return cls(
            name=data["name"],
            description=data.get("description", ""),
            input_schema=data.get("input_schema", {})
        )


class MCPClient:
    """Model Context Protocol client for communicating with MCP server."""

    def __init__(self):
        """Initialize the MCP client."""
        self.session: Optional[ClientSession] = None
        self.logger = logger
        self._running = False
        self._connection_task = None
        self.tools: List[Tool] = []
        self._exit_stack = AsyncExitStack()
        self._cleanup_lock = asyncio.Lock()

    async def connect_to_server(self, server_script_path: str, max_retries: int = 3, retry_delay: float = 2.0):
        """Connect to the MCP server.

        Args:
            server_script_path: Path to the server script
            max_retries: Maximum number of connection retries
            retry_delay: Delay between retries in seconds
        """
        try:
            # Validate server script path
            if not os.path.exists(server_script_path):
                raise ValueError(
                    f"Server script not found at {server_script_path}")

            self.logger.info(
                f"Connecting to MCP server at {server_script_path}")

            # Create server parameters with enhanced configuration
            server_params = StdioServerParameters(
                command=sys.executable,
                args=[server_script_path],
                env=os.environ.copy()
            )

            # Start connection in background task with timeout
            self._running = True
            try:
                self._connection_task = asyncio.create_task(
                    asyncio.wait_for(
                        self._maintain_connection(
                            server_params, max_retries, retry_delay),
                        timeout=60.0  # Overall connection timeout
                    )
                )
                self.logger.info("Started MCP client connection task")
            except asyncio.TimeoutError:
                self.logger.error("MCP connection setup timed out")
                self._running = False
                raise

        except Exception as e:
            self.logger.error(
                f"Failed to connect to MCP server: {str(e)}", exc_info=True)
            self._running = False
            # Don't raise the exception to prevent app startup failure
            self.logger.warning("MCP client will continue without connection")

    async def _maintain_connection(self, server_params: StdioServerParameters, max_retries: int, retry_delay: float):
        """Maintain the connection to the MCP server.

        Args:
            server_params: Server parameters for connection
            max_retries: Maximum number of connection retries
            retry_delay: Delay between retries in seconds
        """
        retry_count = 0

        while self._running and retry_count < max_retries:
            try:
                self.logger.info("Establishing connection to MCP server...")

                # Use AsyncExitStack for proper resource management
                read, write = await self._exit_stack.enter_async_context(stdio_client(server_params))
                self.session = await self._exit_stack.enter_async_context(ClientSession(read, write))

                # Initialize the session
                await self.session.initialize()
                self.logger.info("Connected to MCP server successfully")

                # Initialize tools
                try:
                    await self._initialize_tools()
                except Exception as e:
                    self.logger.error(f"Error initializing tools: {str(e)}")
                    self.tools = []

                # Keep connection alive until cleanup is called
                while self._running:
                    await asyncio.sleep(1)

            except Exception as e:
                self.logger.error(f"Connection error: {str(e)}", exc_info=True)
                retry_count += 1
                self.session = None

                if self._running and retry_count < max_retries:
                    # Wait before retrying
                    self.logger.info(
                        f"Retry attempt {retry_count}/{max_retries} in {retry_delay} seconds...")
                    await asyncio.sleep(retry_delay)
                else:
                    self.tools = []
                    if retry_count >= max_retries:
                        self.logger.error(
                            f"Failed to connect after {max_retries} attempts")
                        break

    async def _initialize_tools(self):
        """Initialize tools from the server."""
        if not self.session:
            raise RuntimeError("No active session to initialize tools")

        # Add retry logic for tool initialization
        max_retries = 3
        retry_delay = 2.0

        for attempt in range(max_retries):
            try:
                self.logger.info(
                    f"Attempting to initialize tools (attempt {attempt+1}/{max_retries})...")
                tools_response = await self.session.list_tools()
                self.logger.info(f"Raw tools response: {tools_response}")

                # Check if tools_response is None
                if tools_response is None:
                    self.logger.warning(
                        f"Tools response is None on attempt {attempt+1}")
                    if attempt < max_retries - 1:
                        self.logger.info(
                            f"Retrying in {retry_delay} seconds...")
                        await asyncio.sleep(retry_delay)
                    continue

                # Check if tools attribute exists and is not empty
                if hasattr(tools_response, 'tools') and tools_response.tools:
                    # Detailed logging about the tools received
                    self.logger.info(
                        f"Received {len(tools_response.tools)} tools from server")
                    for i, tool in enumerate(tools_response.tools):
                        tool_desc = tool.description if tool.description else ""
                        desc_preview = tool_desc[:30] + \
                            "..." if tool_desc else "(no description)"
                        self.logger.info(
                            f"Tool {i+1}: name={tool.name}, desc={desc_preview}")

                    # Create Tool objects
                    self.tools = [
                        Tool(
                            name=t.name,
                            description=t.description or "",
                            input_schema=t.inputSchema
                        ) for t in tools_response.tools
                    ]

                    self.logger.info(
                        f"Successfully initialized {len(self.tools)} tools")

                    # Log the names of the tools
                    tool_names = [t.name for t in self.tools]
                    self.logger.info(f"Available tools: {tool_names}")
                    return  # Success, exit the retry loop
                else:
                    # Log the structure of tools_response for debugging
                    self.logger.warning("No tools found in response.")
                    if hasattr(tools_response, 'tools'):
                        self.logger.warning(
                            f"tools attribute exists but contains {tools_response.tools}")
                    else:
                        self.logger.warning(
                            f"tools attribute does not exist. Response type: {type(tools_response)}")
                        self.logger.warning(
                            f"Response attributes: {dir(tools_response)}")

                    # Reset tools list
                    self.tools = []

                    if attempt < max_retries - 1:
                        self.logger.info(
                            f"Retrying tool initialization in {retry_delay} seconds...")
                        await asyncio.sleep(retry_delay)
            except Exception as e:
                self.logger.error(
                    f"Error initializing tools (attempt {attempt+1}): {str(e)}", exc_info=True)
                self.tools = []
                if attempt < max_retries - 1:
                    self.logger.info(
                        f"Retrying after error in {retry_delay} seconds...")
                    await asyncio.sleep(retry_delay)

        # If all retries fail
        self.logger.error(
            f"Failed to initialize tools after {max_retries} attempts")
        self.tools = []

    async def cleanup(self):
        """Clean up resources properly."""
        async with self._cleanup_lock:
            self.logger.info("Cleaning up MCP client...")
            self._running = False

            if self._connection_task and not self._connection_task.done():
                try:
                    self._connection_task.cancel()
                    await asyncio.wait_for(self._connection_task, timeout=2.0)
                except (asyncio.CancelledError, asyncio.TimeoutError):
                    self.logger.info("Connection task cancelled")
                except Exception as e:
                    self.logger.error(
                        f"Error during connection task cleanup: {str(e)}", exc_info=True)

            try:
                await self._exit_stack.aclose()
            except Exception as e:
                self.logger.error(
                    f"Error during exit stack cleanup: {str(e)}", exc_info=True)

            self.session = None
            self.logger.info("MCP client cleanup complete")

    async def get_tools(self) -> List[Dict[str, Any]]:
        """Get list of available tools from the MCP server."""
        return [tool.to_dict() for tool in self.tools]

    async def invoke_tool(self, tool_name: str, tool_args: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Invoke a tool on the MCP server (alias for call_tool for compatibility).

        Args:
            tool_name: Name of the tool to call
            tool_args: Arguments to pass to the tool

        Returns:
            Dictionary with status and content/error fields
        """
        return await self.call_tool(tool_name, tool_args)

    async def call_tool(self, tool_name: str, tool_args: Optional[Dict[str, Any]] = None, retries: int = 2) -> Dict[str, Any]:
        """Call a tool on the MCP server with retry logic.

        Args:
            tool_name: Name of the tool to call
            tool_args: Arguments to pass to the tool
            retries: Number of retry attempts

        Returns:
            Dictionary with status and content/error fields
        """
        if not self.session:
            self.logger.error("[TOOL_CALL] No active session to call tool")
            return {"status": "error", "error": "No active MCP session"}

        attempt = 0
        last_exception = None

        # Log the tool being called
        self.logger.info(
            f"[TOOL_CALL] Calling tool '{tool_name}' with args: {json.dumps(tool_args or {}, default=str)[:200]}...")

        while attempt <= retries:
            try:
                # Handle authentication properly
                if tool_args and "authorization" in tool_args:
                    auth = tool_args.get("authorization")
                    if not auth or auth == "Bearer undefined" or auth == "Bearer null":
                        self.logger.warning(
                            f"[TOOL_CALL] Missing or invalid authorization token: {auth} - proceeding without authentication")
                        # Remove invalid auth to prevent errors downstream
                        tool_args.pop("authorization")
                    else:
                        # Only show beginning of token for security
                        auth_preview = auth[:20] + \
                            "..." if len(auth) > 20 else auth
                        self.logger.info(
                            f"[TOOL_CALL] Using authorization: {auth_preview}")

                self.logger.info(
                    f"[TOOL_CALL] Executing tool '{tool_name}' (attempt {attempt+1}/{retries+1})")

                # Check if the tool exists in our list of tools
                tool_exists = False
                for tool in self.tools:
                    if tool.name == tool_name:
                        tool_exists = True
                        break

                if not tool_exists:
                    self.logger.warning(
                        f"[TOOL_CALL] Tool '{tool_name}' not found in registered tools. Available tools: {[t.name for t in self.tools]}")

                # Record start time for latency tracking
                start_time = time.time()

                # Call the tool
                self.logger.info(
                    f"[TOOL_CALL] Sending request to server for tool '{tool_name}'")
                result = await self.session.call_tool(tool_name, arguments=tool_args or {})

                # Calculate and log latency
                latency = time.time() - start_time
                self.logger.info(
                    f"[TOOL_CALL] Tool '{tool_name}' call completed in {latency:.3f} seconds")

                # Log successful response
                if isinstance(result, dict):
                    result_preview = str(
                        result)[:1000] + "..." if len(str(result)) > 100 else str(result)
                    self.logger.info(
                        f"[TOOL_CALL] Tool '{tool_name}' returned dict: {result_preview}")

                    # Check for status field if present
                    if "status" in result:
                        self.logger.info(
                            f"[TOOL_CALL] Response status: {result.get('status')}")
                else:
                    result_preview = str(
                        result)[:1000] + "..." if len(str(result)) > 100 else str(result)
                    self.logger.info(
                        f"[TOOL_CALL] Tool '{tool_name}' returned: {result_preview}")

                self.logger.info(
                    f"[TOOL_CALL] Successfully executed tool '{tool_name}'")
                return {
                    "status": "success",
                    "content": result
                }
            except Exception as e:
                attempt += 1
                last_exception = e
                self.logger.error(
                    f"[TOOL_CALL] Error calling tool '{tool_name}' (attempt {attempt}/{retries+1}): {str(e)}")

                if attempt <= retries:
                    self.logger.info(
                        f"[TOOL_CALL] Retrying tool '{tool_name}' in 1.0 seconds...")
                    await asyncio.sleep(1.0)  # Wait before retry

        # Return error if all retries failed
        self.logger.error(
            f"[TOOL_CALL] All attempts to call tool '{tool_name}' failed after {retries+1} tries")
        return {
            "status": "error",
            "error": str(last_exception) if last_exception else "Unknown error"
        }
