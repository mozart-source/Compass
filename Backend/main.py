import asyncio
import sys
import os
import codecs
from data_layer.cache.dashboard_cache import DashboardCache
from data_layer.mongodb.lifecycle import mongodb_lifespan
from data_layer.mongodb.connection import get_mongodb_client
from core.config import settings
from core.mcp_state import set_mcp_client, get_mcp_client, set_mcp_process
from api.ai_routes import router as ai_router
from data_layer.cache.redis_client import redis_client, redis_pubsub_client
from data_layer.cache.pubsub_manager import pubsub_manager
from data_layer.cache.dashboard_cache import DashboardCache
import pathlib
from contextlib import asynccontextmanager
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, FileResponse
from fastapi import FastAPI, Request, Depends, Cookie, HTTPException, APIRouter
from typing import Dict, Any, Optional, AsyncGenerator
import logging
import json
import asyncio
import sys
import codecs
import datetime
import io
from api.focus_routes import router as focus_router
from api.goal_routes import router as goal_router
from api.system_metric_routes import router as system_metric_router
from api.cost_tracking_routes import router as cost_tracking_router
from api.dashboard_routes import dashboard_router
from api.report_routes import router as report_router
from api.websocket import report_ws
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Import WebSocket manager if available
try:
    from api.websocket.dashboard_ws import dashboard_ws_manager, start_background_tasks, stop_background_tasks
except ImportError:
    dashboard_ws_manager = None
    start_background_tasks = None
    stop_background_tasks = None
    logging.getLogger(__name__).warning(
        "WebSocket manager not available, real-time updates will be disabled")

# Set up proper encoding for stdout/stderr
try:
    sys.stdout = codecs.getwriter('utf-8')(sys.stdout.buffer)
    sys.stderr = codecs.getwriter('utf-8')(sys.stderr.buffer)
except (AttributeError, IOError):
    pass


class EmojiSafeFormatter(logging.Formatter):
    """Log formatter that makes emojis and special characters safe for console output."""

    def format(self, record):
        msg = super().format(record)
        # Replace common emojis with text equivalents
        replacements = {
            '✅': '[OK]',
            '❌': '[X]',
            '⚠️': '[WARN]',
            '🔄': '[REFRESH]',
            '🚀': '[ROCKET]',
            '📊': '[CHART]',
            '🔍': '[SEARCH]',
            '🔒': '[LOCK]'
        }

        for emoji, replacement in replacements.items():
            msg = msg.replace(emoji, replacement)
        return msg

# Configure logging with Unicode safety


class EncodingSafeHandler(logging.StreamHandler):
    """Stream handler that handles encoding errors gracefully."""

    def emit(self, record):
        try:
            msg = self.format(record)
            stream = self.stream
            # Use a safer approach to write to the stream
            stream.write(msg + self.terminator)
            self.flush()
        except UnicodeEncodeError:
            # Fall back to ascii with replacement if Unicode fails
            try:
                msg = self.format(record)
                # Replace problematic characters
                safe_msg = msg.encode('ascii', 'replace').decode('ascii')
                stream = self.stream
                stream.write(safe_msg + self.terminator)
                self.flush()
            except Exception:
                self.handleError(record)
        except Exception:
            self.handleError(record)


# Configure the root logger
logger_format = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
formatter = EmojiSafeFormatter(logger_format)

# Clear any existing handlers
root_logger = logging.getLogger()
for handler in root_logger.handlers[:]:
    root_logger.removeHandler(handler)

# Add new safe handlers
console_handler = EncodingSafeHandler(sys.stdout)
console_handler.setFormatter(formatter)

# Ensure the log directory exists
log_dir = "/app/writable/logs"
os.makedirs(log_dir, exist_ok=True)
log_file = os.path.join(log_dir, "compass.log")

file_handler = logging.FileHandler(log_file, encoding='utf-8')
file_handler.setFormatter(formatter)

logging.basicConfig(
    level=logging.INFO,
    handlers=[console_handler, file_handler]
)

logger = logging.getLogger(__name__)
logger.info("Logging initialized with emoji-safe configuration")

# Forward declarations
startup_event = None
shutdown_event = None


class ApplicationLifecycle:
    def __init__(self):
        self.app = None

    async def startup(self):
        """Initialize application resources on startup"""
        try:
            async def handle_dashboard_event(event):
                try:
                    await dashboard_cache.update(event)
                except Exception as e:
                    logger.error(f"Error handling dashboard event: {str(e)}")

            # Start the Redis subscriber in the background for Python backend events
            asyncio.create_task(redis_pubsub_client.subscribe(
                "dashboard_events", handle_dashboard_event))

            # Start subscribers for both Go backend and Notes server
            if hasattr(dashboard_cache, 'start_go_metrics_subscriber'):
                go_subscriber = dashboard_cache.start_go_metrics_subscriber()
                if asyncio.iscoroutine(go_subscriber):
                    await go_subscriber

            if hasattr(dashboard_cache, 'start_notes_metrics_subscriber'):
                notes_subscriber = dashboard_cache.start_notes_metrics_subscriber()
                if asyncio.iscoroutine(notes_subscriber):
                    await notes_subscriber

            logger.info(
                "Started dashboard metrics subscribers for Go backend and Notes server")

            # Initialize WebSocket manager if available
            if dashboard_ws_manager and start_background_tasks:
                # Start the WebSocket ping task to keep connections alive
                await start_background_tasks()
                logger.info(
                    "WebSocket manager initialized and ping task started")
        except Exception as e:
            logger.error(f"Error during startup: {str(e)}")
            raise

    async def shutdown(self):
        """Cleanup application resources on shutdown"""
        try:
            # Close Redis pub/sub client
            await redis_pubsub_client.close()

            # Close dashboard cache
            if dashboard_cache:
                await dashboard_cache.close()

            # Cleanup WebSocket resources if available
            if dashboard_ws_manager and stop_background_tasks:
                await stop_background_tasks()
                logger.info("WebSocket manager resources cleaned up")
        except Exception as e:
            logger.error(f"Error during shutdown: {str(e)}")
            raise


# Create lifecycle manager
lifecycle = ApplicationLifecycle()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[Any, None]:
    """Initialize and manage resources using multiple lifecycle managers."""
    try:
        # Initialize Redis
        logger.info("Testing Redis connection...")
        try:
            await redis_client.ping()
            logger.info("✅ Redis connection successful on database 1")
        except Exception as e:
            logger.error(f"❌ Redis connection failed: {str(e)}")
            raise

        # Pre-load sentence transformer model
        logger.info("Pre-loading sentence transformer model...")
        try:
            import os
            from sentence_transformers import SentenceTransformer
            # Define a writable cache path inside the container.
            cache_path = os.environ.get(
                "SENTENCE_TRANSFORMERS_HOME", "/app/cache/sentence-transformers")
            os.makedirs(cache_path, exist_ok=True)

            logger.info(f"Using sentence transformer cache path: {cache_path}")
            model = SentenceTransformer(
                'all-MiniLM-L6-v2', cache_folder=cache_path)
            app.state.sentence_transformer = model
            logger.info("✅ Sentence transformer model loaded successfully")
        except Exception as e:
            logger.error(f"❌ Failed to load sentence transformer: {str(e)}")
            raise

        # Initialize Atomic Agents
        logger.info("Initializing Atomic Agents framework...")
        try:
            # Configure for GitHub-hosted model integration
            import os
            import openai
            from ai_services.llm.llm_service import LLMService
            from ai_services.adapters.github_model_adapter import GitHubModelAdapter
            from atomic_agents.lib.components.agent_memory import AgentMemory

            # Set default API key for Atomic Agents (for compatibility only)
            if not os.environ.get("OPENAI_API_KEY"):
                logger.warning(
                    "OPENAI_API_KEY not found in environment, using default key for development")
                os.environ["OPENAI_API_KEY"] = settings.openai_api_key

            # Create GitHub model adapter to integrate with Atomic Agents
            llm_service = LLMService()
            github_adapter = GitHubModelAdapter(llm_service)

            # Store adapter and memory in app state for reuse
            app.state.github_adapter = github_adapter
            app.state.agent_memory = AgentMemory()
            app.state.llm_service = llm_service

            # Make these available globally
            from ai_services.agents.base_agent import set_global_llm_service, set_global_github_adapter, set_global_memory
            set_global_llm_service(llm_service)
            set_global_github_adapter(github_adapter)
            set_global_memory(app.state.agent_memory)

            logger.info(
                "✅ Atomic Agents compatible components initialized successfully")
        except Exception as e:
            logger.error(
                f"❌ Failed to initialize Atomic Agents components: {str(e)}")
            logger.warning(
                "Application will continue without Atomic Agents support")

        # Initialize MongoDB directly (not with async with)
        logger.info("Connecting to MongoDB...")
        from data_layer.mongodb.connection import get_mongodb_client, get_async_mongodb_client
        from data_layer.mongodb.lifecycle import init_collections

        # Initialize clients with graceful degradation
        try:
            mongo_client = get_mongodb_client()
            if mongo_client:
                get_async_mongodb_client()
                # Initialize collections
                await init_collections()
                logger.info("✅ MongoDB connection initialized successfully")
            else:
                logger.warning(
                    "⚠️ MongoDB connection failed - continuing without MongoDB")
        except Exception as mongo_error:
            logger.error(
                f"❌ MongoDB initialization failed: {str(mongo_error)}")
            logger.warning(
                "⚠️ Application will continue without MongoDB support")

        if settings.mcp_enabled:
            logger.info("MCP is enabled, initializing server...")
            init_success = await init_mcp_server()
            if init_success:
                logger.info("✅ MCP server initialized successfully")
                # Validate that client is properly stored
                from core.mcp_state import get_mcp_client
                test_client = get_mcp_client()
                if test_client:
                    try:
                        tools = await test_client.get_tools()
                        logger.info(
                            f"✅ MCP client validated with {len(tools)} tools available")
                    except Exception as validation_error:
                        logger.error(
                            f"❌ MCP client validation failed: {str(validation_error)}")
                        # Clear the faulty client
                        from core.mcp_state import set_mcp_client
                        set_mcp_client(None)
                else:
                    logger.error(
                        "❌ MCP client not found in global state after initialization")
            else:
                logger.error("❌ MCP server initialization failed")

        logger.info("Application started successfully")
        # Call lifecycle startup
        await lifecycle.startup()
        yield
    finally:
        logger.info("Shutting down application...")

        # Gracefully shutdown application resources before closing connections
        await lifecycle.shutdown()

        if settings.mcp_enabled:
            await cleanup_mcp()

        try:
            # Close MongoDB connections
            from data_layer.mongodb.connection import close_mongodb_connections
            await close_mongodb_connections()
            logger.info("MongoDB connections closed")

            await redis_client.close()
            logger.info("Redis connection closed")
        except Exception as e:
            logger.error(f"Error closing connections: {str(e)}")

        logger.info("All resources cleaned up")

# Create FastAPI app with lifespan
app = FastAPI(
    title="COMPASS Backend API",
    description="API for the COMPASS productivity platform",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,  # Use dynamic CORS origins from settings
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE",
                   "OPTIONS", "PATCH"],  # Specify allowed methods
    allow_headers=[
        "Authorization",
        "Content-Type",
        "X-Requested-With",
        "X-Organization-ID",
        "x-organization-id",
        "Accept",
        "Origin",
        "Accept-Encoding",
        "Content-Encoding",
        "Cache-Control",
        "Pragma"
    ],
)

# Health check router for the /api/v1 prefix
health_router = APIRouter()


@health_router.get("/health")
async def prefixed_health_check():
    return await health_check()


# Include routers
app.include_router(health_router, prefix="/api/v1")
app.include_router(ai_router, prefix="/api/v1", tags=["AI"])
app.include_router(focus_router, prefix="/api/v1", tags=["Focus"])
app.include_router(goal_router, prefix="/api/v1", tags=["Goals"])
app.include_router(system_metric_router, prefix="/api/v1",
                   tags=["System Metrics"])
app.include_router(cost_tracking_router, prefix="/api/v1",
                   tags=["Cost Tracking"])
app.include_router(dashboard_router, prefix="/api/v1", tags=["Dashboard"])
app.include_router(report_router, prefix="/api/v1", tags=["Reports"])
app.include_router(report_ws.router, prefix="/api/v1/ws",
                   tags=["Reports WebSocket"])

# Include WebSocket routes if available
try:
    from api.websocket.routes import router as websocket_router
    app.include_router(websocket_router, prefix="/api/v1")
    logger.info("WebSocket routes included successfully")
except ImportError as e:
    logger.warning(f"Failed to include WebSocket routes: {str(e)}")

# Mount static files directory only if it exists
static_dir = pathlib.Path("static")
if static_dir.exists() and static_dir.is_dir():
    logger.info(f"Mounting static files directory: {static_dir}")
    app.mount("/static", StaticFiles(directory="static"), name="static")
else:
    logger.warning(
        f"Static directory '{static_dir}' does not exist - not mounting static files")
    # Create the directory to prevent future errors if needed
    try:
        static_dir.mkdir(exist_ok=True)
        logger.info(f"Created static directory: {static_dir}")
    except Exception as e:
        logger.warning(f"Could not create static directory: {str(e)}")

# Create global instance
dashboard_cache = DashboardCache()


async def revalidate_mcp_client():
    """Re-validate MCP client and update global state if tools become available."""
    from core.mcp_state import get_mcp_client, set_mcp_client

    max_attempts = 10
    for attempt in range(max_attempts):
        await asyncio.sleep(10)  # Wait 10 seconds between attempts

        client = get_mcp_client()
        if not client:
            logger.info(
                f"Re-validation attempt {attempt + 1}: No MCP client available")
            continue

        try:
            tools = await client.get_tools()
            if len(tools) > 0:
                logger.info(
                    f"✅ MCP client re-validation successful: {len(tools)} tools now available")

                # Ensure the client is properly stored in global state
                set_mcp_client(client)
                logger.info("✅ MCP client re-stored in global state")

                # Test global state access
                test_client = get_mcp_client()
                if test_client:
                    logger.info("✅ MCP client verified in global state")
                else:
                    logger.error(
                        "❌ MCP client not found in global state after re-storing")

                return
            else:
                logger.info(
                    f"Re-validation attempt {attempt + 1}: Still no tools available")
        except Exception as e:
            logger.warning(
                f"Re-validation attempt {attempt + 1} failed: {str(e)}")

    logger.error("❌ MCP client re-validation failed after all attempts")


async def init_mcp_server():
    """Initialize the MCP server."""
    logger.info("Initializing MCP server integration")
    env = os.environ.copy()

    # For debugging
    logger.info(f"PYTHONPATH: {sys.path}")

    # MCP server script
    server_script = os.path.join(os.path.dirname(
        os.path.abspath(__file__)), "mcp_py", "server.py")
    logger.info(f"Starting MCP server: {server_script}")

    # Check if server script exists
    if not os.path.exists(server_script):
        logger.error(f"MCP server script not found: {server_script}")
        return False

    try:
        # Check if we're on Windows
        if sys.platform == 'win32':
            # Windows-specific approach - use subprocess directly instead of asyncio.create_subprocess_exec
            import subprocess
            process = subprocess.Popen(
                [sys.executable, server_script],
                env=env,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )

            # Store the process in global state so it can be terminated later
            set_mcp_process(process)

            # Wait longer for server to start and validate it's running
            await asyncio.sleep(5)

            # Check if process is still running
            if process.poll() is not None:
                # Process has terminated, get error output
                stdout, stderr = process.communicate()
                logger.error(f"MCP server process terminated unexpectedly")
                logger.error(f"STDOUT: {stdout}")
                logger.error(f"STDERR: {stderr}")
                return False

            logger.info(
                "MCP server process is running, attempting client connection...")

            # Set up MCP client with retry logic
            from mcp_py.client import MCPClient
            client = MCPClient()

            max_retries = 3
            for attempt in range(max_retries):
                try:
                    logger.info(
                        f"MCP client connection attempt {attempt + 1}/{max_retries}")
                    # This call needs to be awaited even on windows
                    await client.connect_to_server(server_script)

                    # Test the connection by getting tools
                    tools = await client.get_tools()
                    logger.info(
                        f"MCP client connected successfully with {len(tools)} tools")

                    # Store the client in global state for other components to use
                    set_mcp_client(client)

                    # If no tools available initially, schedule a background re-validation
                    if len(tools) == 0:
                        logger.warning(
                            "MCP client connected but no tools available initially, will retry in background")
                        asyncio.create_task(revalidate_mcp_client())

                    return True

                except Exception as connect_error:
                    logger.warning(
                        f"MCP client connection attempt {attempt + 1} failed: {str(connect_error)}")
                    if attempt < max_retries - 1:
                        await asyncio.sleep(2)  # Wait before retry
                    else:
                        logger.error(
                            f"All MCP client connection attempts failed: {str(connect_error)}")
                        return False
        else:
            # Original code for Unix-like systems
            process = await asyncio.create_subprocess_exec(
                sys.executable, server_script,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                env=env
            )

            set_mcp_process(process)

            # Give the server a moment to start
            await asyncio.sleep(3)

            # Set up MCP client with retry logic
            from mcp_py.client import MCPClient
            client = MCPClient()

            max_retries = 3
            for attempt in range(max_retries):
                try:
                    logger.info(
                        f"MCP client connection attempt {attempt + 1}/{max_retries}")
                    await client.connect_to_server(server_script)

                    # Test the connection
                    tools = await client.get_tools()
                    logger.info(
                        f"MCP client connected successfully with {len(tools)} tools")

                    # Store the client in global state for other components to use
                    set_mcp_client(client)

                    # If no tools available initially, schedule a background re-validation
                    if len(tools) == 0:
                        logger.warning(
                            "MCP client connected but no tools available initially, will retry in background")
                        asyncio.create_task(revalidate_mcp_client())

                    return True

                except Exception as connect_error:
                    logger.warning(
                        f"MCP client connection attempt {attempt + 1} failed: {str(connect_error)}")
                    if attempt < max_retries - 1:
                        await asyncio.sleep(2)  # Wait before retry
                    else:
                        logger.error(
                            f"All MCP client connection attempts failed: {str(connect_error)}")
                        return False

    except Exception as e:
        logger.error(f"Error initializing MCP server: {str(e)}", exc_info=True)
        return False

    return False


async def cleanup_mcp():
    """Clean up MCP client and server resources."""
    from core.mcp_state import get_mcp_client, get_mcp_process

    # Close MCP client if it exists
    mcp_client = get_mcp_client()
    if mcp_client:
        logger.info("Closing MCP client connection")
        try:
            await mcp_client.cleanup()
        except Exception as e:
            logger.error(f"Error cleaning up MCP client: {str(e)}")

    # Terminate MCP server process if it exists
    mcp_process = get_mcp_process()
    if mcp_process:
        logger.info("Terminating MCP server process")
        try:
            # Handle both subprocess.Popen (Windows) and asyncio subprocess (Unix)
            if hasattr(mcp_process, 'terminate'):  # subprocess.Popen
                mcp_process.terminate()
                # Wait for process to terminate
                try:
                    mcp_process.wait(timeout=5)
                except:
                    # Force kill if timeout
                    if hasattr(mcp_process, 'kill'):
                        mcp_process.kill()
            else:  # asyncio subprocess
                mcp_process.terminate()
                # Wait for process to terminate
                try:
                    await asyncio.wait_for(mcp_process.wait(), timeout=5)
                except asyncio.TimeoutError:
                    # Force kill if timeout
                    mcp_process.kill()

        except Exception as e:
            logger.error(f"Error terminating MCP server process: {str(e)}")

# Root endpoint


@app.get("/")
async def root():
    return {"message": "Welcome to COMPASS Backend", "version": "1.0.0"}

# Health check endpoint


@app.get("/health")
async def health_check():
    """Health check endpoint for Docker healthcheck."""
    try:
        # Check MongoDB connection
        client = get_mongodb_client()
        if client:
            db = client.admin
            server_info = db.command("ping")
            mongodb_ok = server_info.get("ok") == 1.0
        else:
            mongodb_ok = False
    except Exception as e:
        logger.error(f"MongoDB health check error: {str(e)}")
        mongodb_ok = False

    # Check Redis connection
    try:
        redis_ok = await redis_client.ping()
        logger.info("Redis health check passed")
    except Exception as e:
        logger.error(f"Redis health check error: {str(e)}")
        redis_ok = False

    # Check WebSocket server status if available
    websocket_ok = False
    if dashboard_ws_manager:
        websocket_ok = True
        logger.info("WebSocket server health check passed")

    return {
        "status": "healthy" if (mongodb_ok and redis_ok) else "degraded",
        "mongodb": mongodb_ok,
        "redis": redis_ok,
        "redis_db": 1,  # Show which Redis DB we're using
        "websocket": websocket_ok,
        "timestamp": datetime.datetime.utcnow().isoformat()
    }

# Global exception handler


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Global exception handler for unhandled exceptions."""
    logger.exception(f"Unhandled exception: {str(exc)}")
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"},
    )

if __name__ == "__main__":
    import uvicorn

    # Check if HTTPS is enabled in settings
    if settings.use_https:
        logger.info(
            f"Starting server with HTTPS on {settings.api_host}:{settings.api_port}")
        logger.info(f"Using certificate: {settings.https_cert_file}")
        logger.info(f"Using key: {settings.https_key_file}")

        uvicorn.run(
            "main:app",
            host=settings.api_host,
            port=settings.api_port,
            reload=True,
            ssl_keyfile=settings.https_key_file,
            ssl_certfile=settings.https_cert_file
        )
    else:
        logger.info(
            f"Starting server with HTTP on {settings.api_host}:{settings.api_port}")
        uvicorn.run("main:app", host=settings.api_host,
                    port=settings.api_port, reload=True)
