import asyncio
import aiohttp
import json
import time
from datetime import datetime, timedelta
from uuid import UUID
from typing import Optional, Dict, Any, List, Tuple
from app.schemas.dashboard_metrics import DashboardMetrics, DailyFocusItem
from data_layer.repos.focus_repo import FocusSessionRepository, FocusSettingsRepository
from data_layer.repos.goal_repo import GoalRepository, Goal
from data_layer.repos.system_metric_repo import SystemMetricRepository
from data_layer.repos.ai_model_repo import ModelUsageRepository, ModelUsage
from data_layer.repos.cost_tracking_repo import CostTrackingRepository
from core.config import settings
from data_layer.cache.redis_client import redis_client, redis_pubsub_client
import logging
import os
from data_layer.cache.pubsub_manager import PubSubManager

# Import WebSocket manager
try:
    from api.websocket.dashboard_ws import dashboard_ws_manager
except ImportError:
    dashboard_ws_manager = None
    logging.getLogger(__name__).warning(
        "WebSocket manager not available, real-time updates will be disabled")

# Define Go backend dashboard event types


class events:
    DashboardEventMetricsUpdate = "metrics_update"
    DashboardEventCacheInvalidate = "cache_invalidate"

# Custom error classes for better error handling


class DashboardError(Exception):
    def __init__(self, message: str, error_type: str, details: Optional[Dict] = None):
        self.message = message
        self.error_type = error_type
        self.details = details or {}
        super().__init__(self.message)


class DashboardMetricsError(DashboardError):
    pass


class CircuitBreaker:
    """Circuit breaker pattern implementation to prevent cascading failures"""

    def __init__(self, failure_threshold=5, reset_timeout=60, name="default"):
        self._failure_count = 0
        self._failure_threshold = failure_threshold
        self._reset_timeout = reset_timeout
        self._last_failure_time = 0
        self._is_open = False
        self.name = name

    @property
    def is_open(self):
        """Check if circuit is open"""
        return self._is_open

    @property
    def failure_count(self):
        """Get current failure count"""
        return self._failure_count

    @property
    def last_failure_time(self):
        """Get last failure timestamp"""
        return self._last_failure_time

    async def execute(self, func, *args, **kwargs):
        if self._is_open:
            if time.time() - self._last_failure_time > self._reset_timeout:
                self._is_open = False
            else:
                raise DashboardError(
                    "Circuit breaker is open",
                    "circuit_open",
                    {"reset_in": self._reset_timeout -
                        (time.time() - self._last_failure_time)}
                )

        try:
            result = await func(*args, **kwargs)
            self._failure_count = 0
            return result
        except Exception as e:
            self._failure_count += 1
            if self._failure_count >= self._failure_threshold:
                self._is_open = True
                self._last_failure_time = time.time()
            raise


logger = logging.getLogger(__name__)

focus_repo = FocusSessionRepository()
goal_repo = GoalRepository()
system_repo = SystemMetricRepository()
ai_usage_repo = ModelUsageRepository()
cost_repo = CostTrackingRepository()


class DashboardEvent:
    def __init__(self, event_type: str, user_id: str, entity_id: str, details: Optional[Dict[str, Any]] = None):
        self.event_type = event_type
        self.user_id = user_id
        self.entity_id = entity_id
        self.timestamp = datetime.utcnow()
        self.details = details or {}

    def to_dict(self) -> Dict[str, Any]:
        return {
            "event_type": self.event_type,
            "user_id": self.user_id,
            "entity_id": self.entity_id,
            "timestamp": self.timestamp.isoformat(),
            "details": self.details
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'DashboardEvent':
        return cls(
            event_type=data["event_type"],
            user_id=data["user_id"],
            entity_id=data["entity_id"],
            details=data.get("details")
        )


class DashboardCache:
    def __init__(self):
        self.go_backend_url = settings.GO_BACKEND_URL
        self.redis_client = redis_client
        self.pubsub_manager = PubSubManager()
        self.is_subscribed = False
        self.notes_server_url = settings.NOTES_SERVER_URL
        self.session = None
        self.subscriber_task = None
        # Store reference to WebSocket manager if available
        self.ws_manager = dashboard_ws_manager
        # In-memory cache for faster access
        self._memory_cache = {}
        # Cache lock to prevent race conditions
        self._cache_lock = asyncio.Lock()
        # Circuit breakers for external services
        self._go_backend_circuit = CircuitBreaker(
            failure_threshold=3, reset_timeout=30)
        self._notes_server_circuit = CircuitBreaker(
            failure_threshold=3, reset_timeout=30)
        # Metrics collection
        self._metrics = {
            "cache_hits": 0,
            "cache_misses": 0,
            "memory_cache_hits": 0,
            "memory_cache_misses": 0,
            "fetch_times": [],
            "errors": 0
        }

        # Real-time update configuration
        self.config = {
            "enable_realtime_updates": os.getenv("ENABLE_REALTIME_UPDATES", "true").lower() == "true",
            # Increase throttle to reduce message frequency (0.3s = 300ms)
            "update_throttle_seconds": float(os.getenv("UPDATE_THROTTLE_SECONDS", "0.3")),
            # Enable batching to reduce message frequency
            "batch_updates": os.getenv("BATCH_UPDATES", "true").lower() == "true",
            "quiet_mode": os.getenv("DASHBOARD_QUIET_MODE", "false").lower() == "true"
        }

        # Track last update times to throttle frequent updates
        self._last_update_times = {}
        self._pending_updates = {}

        # Message deduplication - track recent events to prevent duplicates
        self._recent_events = {}  # {user_id: {event_key: timestamp}}
        # Increased deduplication window to avoid duplicate messages (1.0s instead of 0.5s)
        self._dedup_window = float(os.getenv("DASHBOARD_DEDUP_WINDOW", "1.0"))

        # Log configuration on startup
        if not self.config["quiet_mode"]:
            logger.info(
                f"Dashboard cache configuration: "
                f"realtime={self.config['enable_realtime_updates']}, "
                f"throttle={self.config['update_throttle_seconds']}s, "
                f"batch={self.config['batch_updates']}, "
                f"quiet={self.config['quiet_mode']}, "
                f"dedup_window={self._dedup_window}s"
            )

    async def _get_session(self):
        if self.session is None:
            self.session = aiohttp.ClientSession()
        return self.session

    async def _make_request(self, url, method, headers=None, data=None, timeout=10.0):
        """Make HTTP request to external service with timeout"""
        session = await self._get_session()
        try:
            logger.debug(
                f"Making {method} request to {url} with headers: {headers}")
            async with session.request(method, url, headers=headers, json=data, timeout=timeout) as response:
                logger.debug(f"Response status: {response.status} for {url}")

                if response.status == 200:
                    return await response.json()
                elif response.status == 401:
                    logger.error(
                        f"Authentication failed for {url}. Status: {response.status}")
                    error_text = await response.text()
                    logger.error(f"Response body: {error_text}")
                    raise DashboardError(
                        f"Authentication failed for {url}: {error_text}",
                        "auth_error",
                        {"status_code": response.status, "url": url}
                    )
                else:
                    error_text = await response.text()
                    logger.warning(
                        f"Request to {url} failed with status {response.status}: {error_text}")
                    raise DashboardError(
                        f"Request failed with status {response.status}: {error_text}",
                        "http_error",
                        {"status_code": response.status, "url": url}
                    )
        except asyncio.TimeoutError:
            logger.warning(
                f"Request to {url} timed out after {timeout} seconds")
            raise DashboardError(
                f"Request to {url} timed out", "timeout_error")
        except Exception as e:
            logger.error(f"Error making request to {url}: {str(e)}")
            raise DashboardError(
                f"Error making request to {url}: {str(e)}", "request_error")

    async def _get_focus_metrics(self, user_id: str):
        """Fetch focus metrics asynchronously"""
        try:
            focus_repo = FocusSessionRepository()
            settings_repo = FocusSettingsRepository()
            # Run sync operation in thread pool to make it async
            import asyncio
            loop = asyncio.get_event_loop()
            stats = await loop.run_in_executor(None, focus_repo.get_stats, user_id, 30)

            # Get user's focus settings
            user_settings = await loop.run_in_executor(None, settings_repo.get_user_settings, user_id)

            # Format the focus stats for consistent dashboard metrics structure
            focus_metrics = {
                "total_focus_seconds": stats.get("total_focus_seconds", 0),
                "streak": stats.get("streak", 0),
                "longest_streak": stats.get("longest_streak", 0),
                "sessions": stats.get("sessions", 0),
                "daily_target_seconds": user_settings.daily_target_seconds,
                # Add a daily breakdown for visualization
                "daily_breakdown": self._generate_daily_focus_breakdown(user_id)
            }

            return focus_metrics
        except Exception as e:
            logger.error(f"Error fetching focus metrics: {str(e)}")
            return {
                "total_focus_seconds": 0,
                "streak": 0,
                "longest_streak": 0,
                "sessions": 0,
                "daily_target_seconds": 14400,  # Default 4 hours
                "daily_breakdown": []
            }

    def _generate_daily_focus_breakdown(self, user_id: str):
        """Generate daily focus breakdown for the last 7 days"""
        try:
            from datetime import datetime, timedelta
            from data_layer.repos.focus_repo import FocusSessionRepository

            repo = FocusSessionRepository()

            # Get the current date and calculate the start date (7 days ago)
            today = datetime.now()
            start_date = today - timedelta(days=6)

            # Initialize days with proper format for the last 7 days
            days = []
            for i in range(7):
                day_date = start_date + timedelta(days=i)
                # Short day name (Mon, Tue, etc.)
                day_name = day_date.strftime("%a")
                days.append({
                    "day": day_name,
                    "minutes": 0  # Default to 0, will be updated below
                })

            # Get focus sessions for the last 7 days
            since = start_date.replace(
                hour=0, minute=0, second=0, microsecond=0)
            sessions = repo.find_many({
                "user_id": user_id,
                "start_time": {"$gte": since},
                "status": "completed"
            })

            # Calculate total focus minutes for each day
            day_totals = {}
            for session in sessions:
                if session.start_time and session.duration:
                    session_date = session.start_time.date()
                    day_name = session_date.strftime("%a")

                    # Convert seconds to minutes
                    minutes = session.duration / 60

                    # Add to daily total
                    if day_name in day_totals:
                        day_totals[day_name] += minutes
                    else:
                        day_totals[day_name] = minutes

            # Update the days list with actual focus minutes
            for day in days:
                if day["day"] in day_totals:
                    day["minutes"] = int(day_totals[day["day"]])

            return days
        except Exception as e:
            logger.error(f"Error generating daily focus breakdown: {str(e)}")
            # Fallback to static data
            return [
                {"day": "Mon", "minutes": 40},
                {"day": "Tue", "minutes": 65},
                {"day": "Wed", "minutes": 45},
                {"day": "Thu", "minutes": 80},
                {"day": "Fri", "minutes": 55},
                {"day": "Sat", "minutes": 85},
                {"day": "Sun", "minutes": 60}
            ]

    async def _get_goal_metrics_async(self, user_id: str):
        """Fetch goal metrics asynchronously"""
        try:
            goal_repo = GoalRepository()
            # Run sync operation in thread pool to make it async
            import asyncio
            loop = asyncio.get_event_loop()
            goals = await loop.run_in_executor(None, goal_repo.find_by_user, user_id)
            return self._calculate_goal_metrics(goals)
        except Exception as e:
            logger.error(f"Error fetching goal metrics: {str(e)}")
            return {}

    def _calculate_goal_metrics(self, goals: List[Goal]) -> Dict[str, Any]:
        """Calculate goal metrics from goals list"""
        if not goals:
            return {"total": 0, "completed": 0}

        total = len(goals)
        completed = sum(1 for goal in goals if getattr(
            goal, 'completed', False))

        return {
            "total": total,
            "completed": completed
        }

    async def _get_system_metrics(self, user_id: str):
        """Fetch system metrics asynchronously"""
        try:
            system_metrics_repo = SystemMetricRepository()
            # Run sync operation in thread pool to make it async
            import asyncio
            loop = asyncio.get_event_loop()
            metrics = await loop.run_in_executor(None, system_metrics_repo.aggregate_metrics, user_id, "daily")
            return metrics
        except Exception as e:
            logger.error(f"Error fetching system metrics: {str(e)}")
            return {}

    async def _get_ai_usage_metrics(self, user_id: str):
        """Fetch AI usage metrics asynchronously"""
        try:
            model_usage_repo = ModelUsageRepository()
            # Run sync operation in thread pool to make it async
            import asyncio
            loop = asyncio.get_event_loop()
            usage = await loop.run_in_executor(None, model_usage_repo.get_usage_by_user, user_id, 100)
            return usage
        except Exception as e:
            logger.error(f"Error fetching AI usage metrics: {str(e)}")
            return {}

    async def get_metrics(self, user_id: str, token: str = ""):
        start_time = time.time()

        try:
            # Check memory cache first (1-second threshold)
            if user_id in self._memory_cache:
                metrics, timestamp = self._memory_cache[user_id]
                if time.time() - timestamp < 1:  # 1 second threshold
                    logger.debug(
                        f"Memory cache hit for dashboard metrics: {user_id}")
                    self._metrics["cache_hits"] += 1
                    return metrics

            # Check Redis cache
            cache_key = f"dashboard:metrics:{user_id}"
            cached_metrics = await redis_client.get(cache_key)

            if cached_metrics:
                logger.debug(
                    f"Redis cache hit for dashboard metrics: {user_id}")
                self._metrics["cache_hits"] += 1
                try:
                    metrics = json.loads(cached_metrics)
                    # Update memory cache
                    self._memory_cache[user_id] = (metrics, time.time())
                    return metrics
                except json.JSONDecodeError:
                    logger.warning(
                        f"Failed to decode cached metrics for user {user_id}")
                    # Continue to fetch fresh metrics

            # Cache miss, fetch fresh metrics
            self._metrics["cache_misses"] += 1
            logger.debug(f"Cache miss for dashboard metrics: {user_id}")

            metrics = await self._fetch_all_metrics(user_id, token)

            # Cache the results
            if metrics is not None:
                await self._cache_metrics(user_id, metrics)
                # Update memory cache
                self._memory_cache[user_id] = (metrics, time.time())
            else:
                logger.error(f"Cannot cache null metrics for user {user_id}")

            # Ensure we're subscribed to Go backend events
            if not self.is_subscribed:
                await self.start_go_metrics_subscriber()

            # Record fetch time
            fetch_time = (time.time() - start_time) * 1000  # Convert to ms
            self._metrics["fetch_times"].append(fetch_time)

            return metrics

        except Exception as e:
            self._metrics["errors"] += 1
            logger.error(f"Error getting dashboard metrics: {str(e)}")
            raise DashboardMetricsError(
                str(e),
                "fetch_error",
                {"user_id": user_id}
            )

    async def _fetch_all_metrics(self, user_id: str, token: str = ""):
        """Fetch all metrics from different services"""
        try:
            # Initialize headers with token if provided
            headers = {"Authorization": f"Bearer {token}"} if token else {}
            logger.debug(f"Fetching all metrics for user {user_id}")

            # Fetch metrics from different services concurrently
            go_metrics, notes_metrics, focus, goals, system, ai_usage, cost = await asyncio.gather(
                self._get_go_backend_metrics(user_id, headers),
                self._get_notes_server_metrics(user_id, headers),
                self._get_focus_metrics(user_id),
                self._get_goal_metrics_async(user_id),
                self._get_system_metrics(user_id),
                self._get_ai_usage_metrics(user_id),
                self._get_cost_metrics(user_id),
                return_exceptions=True
            )

            # Extract Go backend metrics
            habits = None
            tasks = None
            todos = None
            calendar = None
            user_metrics = None
            daily_timeline = None
            habit_heatmap = None
            timestamp = datetime.utcnow()

            if not isinstance(go_metrics, Exception) and isinstance(go_metrics, dict):
                data = go_metrics.get("data", {})
                if isinstance(data, dict):
                    habits = data.get("habits")
                    tasks = data.get("tasks")
                    todos = data.get("todos")
                    calendar = data.get("calendar")
                    user_metrics = data.get("user")
                    daily_timeline = data.get("daily_timeline")
                    habit_heatmap = data.get("habit_heatmap")
                    timestamp = data.get("timestamp", timestamp)
                    logger.info(
                        f"Successfully extracted Go backend metrics for user {user_id}: habits={habits is not None}, tasks={tasks is not None}, todos={todos is not None}, calendar={calendar is not None}, user={user_metrics is not None}, timeline={daily_timeline is not None}, heatmap={habit_heatmap is not None}")
                    logger.debug(
                        f"Go backend response data: {json.dumps(data, indent=2)}")
                else:
                    logger.error(
                        f"Invalid data structure in Go backend response: {data}")
            else:
                logger.error(
                    f"Failed to get Go backend metrics for user {user_id}: {go_metrics}")
                logger.error(f"   Exception type: {type(go_metrics)}")
                if isinstance(go_metrics, Exception):
                    logger.error(f"   Exception details: {str(go_metrics)}")

            # Extract Notes server metrics
            mood = None
            notes = None
            journals = None

            if not isinstance(notes_metrics, Exception) and isinstance(notes_metrics, dict):
                # Handle both old and new response formats
                if "data" in notes_metrics and notes_metrics.get("success", True):
                    data = notes_metrics["data"]
                    mood = data.get("mood")
                    notes = data.get("notes")
                    journals = data.get("journals")
                else:
                    # Fallback to old format
                    mood = notes_metrics.get("mood")
                    notes = notes_metrics.get("notes")
                    journals = notes_metrics.get("journals")

                logger.debug(
                    f"Extracted Notes server metrics: mood={mood is not None}, notes={notes is not None}, journals={journals is not None}")
            else:
                logger.error(
                    f"Failed to get Notes server metrics: {notes_metrics}")

            # Handle exceptions in local metrics
            focus = None if isinstance(focus, Exception) else focus
            goals = None if isinstance(goals, Exception) else goals
            system = None if isinstance(system, Exception) else system
            ai_usage = None if isinstance(ai_usage, Exception) else ai_usage
            cost = None if isinstance(cost, Exception) else cost

            # Combine all metrics
            metrics = {
                "habits": habits,
                "calendar": calendar,
                "focus": focus,
                "mood": mood,
                "ai_usage": ai_usage,
                "system_metrics": system,
                "goals": goals,
                "tasks": tasks,
                "todos": todos,
                "user": user_metrics,
                "notes": notes,
                "journals": journals,
                "cost": cost,
                "daily_timeline": daily_timeline,
                "habit_heatmap": habit_heatmap,
                "timestamp": timestamp
            }

            # Validate metrics before returning
            if all(v is None for v in metrics.values()):
                logger.error(f"All metrics are null for user {user_id}")
                return None

            # Log metrics summary without full JSON serialization to avoid coroutine issues
            metrics_summary = {k: v is not None for k, v in metrics.items()}
            logger.debug(
                f"Combined metrics for user {user_id}: {metrics_summary}")
            return metrics

        except Exception as e:
            logger.error(
                f"Error fetching all metrics: {str(e)}", exc_info=True)
            raise DashboardError(
                f"Error fetching metrics: {str(e)}", "fetch_error")

    async def _get_go_backend_metrics(self, user_id: str, headers: dict):
        """Fetch metrics from Go backend"""
        try:
            url = f"{settings.GO_BACKEND_URL}/api/dashboard/metrics"

            # Add service-to-service headers for Go backend authentication bypass
            service_headers = headers.copy() if headers else {}
            service_headers.update({
                "X-Service-Call": "true",
                "X-Internal-Service": "python-backend",
                "User-Agent": "python-backend-aiohttp/dashboard-service"
            })

            logger.info(
                f"🚀 Fetching Go backend metrics from {url} for user {user_id}")
            logger.debug(f"Service headers: {service_headers}")

            # Use circuit breaker for resilience
            response = await self._go_backend_circuit.execute(
                self._make_request, url, "GET", headers=service_headers
            )

            logger.info(
                f"📥 Go backend response received for user {user_id}: {type(response)}")

            if not response:
                logger.error(
                    f"Empty response from Go backend for user {user_id}")
                return None

            if not isinstance(response, dict):
                logger.error(
                    f"Invalid response type from Go backend: {type(response)}")
                return None

            if "data" not in response:
                logger.error(
                    f"Missing 'data' field in Go backend response: {response}")
                return None

            logger.debug(
                f"Successfully fetched Go backend metrics for user {user_id}")
            return response
        except Exception as e:
            logger.error(
                f"Error fetching Go backend metrics: {str(e)}", exc_info=True)
            # If it's an authentication error, log it specifically
            if "401" in str(e) or "unauthorized" in str(e).lower():
                logger.error(
                    f"Authentication error when calling Go backend. Token may be invalid or expired.")
            return None

    async def _get_notes_server_metrics(self, user_id: str, headers: dict):
        """Fetch metrics from Notes server"""
        try:
            url = f"{settings.NOTES_SERVER_URL}/api/dashboard/metrics"
            logger.debug(
                f"Fetching Notes server metrics from {url} with headers: {headers}")

            # Use circuit breaker for resilience
            response = await self._notes_server_circuit.execute(
                self._make_request, url, "GET", headers=headers
            )

            if not response:
                logger.error(
                    f"Empty response from Notes server for user {user_id}")
                return None

            if not isinstance(response, dict):
                logger.error(
                    f"Invalid response type from Notes server: {type(response)}")
                return None

            logger.debug(
                f"Successfully fetched Notes server metrics for user {user_id}")
            return response
        except Exception as e:
            logger.error(
                f"Error fetching Notes server metrics: {str(e)}", exc_info=True)
            # If it's an authentication error, log it specifically
            if "401" in str(e) or "unauthorized" in str(e).lower():
                logger.error(
                    f"Authentication error when calling Notes server. Token may be invalid or expired.")
            return None

    async def _get_cost_metrics(self, user_id: str):
        """Fetch cost metrics"""
        try:
            now = datetime.utcnow()
            last_30 = now - timedelta(days=30)
            cost_repo = CostTrackingRepository()
            # This method is already async, so call it directly
            return await cost_repo.get_user_cost_summary(user_id, last_30, now)
        except Exception as e:
            logger.error(f"Error fetching cost metrics: {str(e)}")
            return None

    async def _cache_metrics(self, user_id: str, metrics: dict):
        """Cache metrics with TTL and update memory cache"""
        cache_key = f"dashboard:metrics:{user_id}"
        # Cache for 5 minutes in Redis
        await redis_client.set(cache_key, json.dumps(metrics), ex=300)
        # Update memory cache
        self._memory_cache[user_id] = (metrics, time.time())
        logger.debug(f"Cached dashboard metrics for user {user_id}")

    async def update_metric(self, user_id: str, metric_type: str, value: Any):
        """Update a specific metric in the cache without invalidating the entire cache"""
        async with self._cache_lock:
            # Get current metrics
            cache_key = f"dashboard:metrics:{user_id}"
            cached_metrics = await redis_client.get(cache_key)

            if not cached_metrics:
                logger.warning(
                    f"Cannot update metric {metric_type} for user {user_id}: cache miss")
                return False

            try:
                current_metrics = json.loads(cached_metrics)
                # Update specific metric
                current_metrics[metric_type] = value

                # Update Redis cache
                await redis_client.set(cache_key, json.dumps(current_metrics), ex=300)

                # Update memory cache
                self._memory_cache[user_id] = (current_metrics, time.time())

                logger.debug(
                    f"Updated metric {metric_type} for user {user_id}")
                return True
            except (json.JSONDecodeError, KeyError) as e:
                logger.error(f"Error updating metric {metric_type}: {str(e)}")
                return False

    async def invalidate_cache(self, user_id: str):
        """Invalidate the cache for a specific user"""
        cache_key = f"dashboard:metrics:{user_id}"
        # Remove from Redis cache
        await redis_client.delete(cache_key)
        # Remove from memory cache
        if user_id in self._memory_cache:
            del self._memory_cache[user_id]
        logger.info(f"Invalidated dashboard cache for user {user_id}")

        # Create an event to notify subscribers
        event = DashboardEvent(
            event_type="cache_invalidate",
            user_id=user_id,
            entity_id="",
            details={"timestamp": datetime.utcnow().isoformat()}
        )
        await self._notify_subscribers(event)

    async def update(self, event: DashboardEvent):
        # Handle real-time updates
        if event.event_type == "dashboard_update":
            # Invalidate cache for this user
            cache_key = f"dashboard:metrics:{event.user_id}"
            await redis_client.delete(cache_key)
            logger.info(
                f"Invalidated dashboard cache for user {event.user_id} due to event {event.event_type}")

            # Notify subscribers
            await self._notify_subscribers(event)

    async def _should_throttle_update(self, event: DashboardEvent) -> bool:
        """Check if we should throttle this update based on frequency"""
        # For immediate user actions (like habit completion), don't throttle
        if event.event_type == "cache_invalidate":
            action = event.details.get('action', '')
            # Don't throttle direct user actions for better UX
            if any(action_type in action for action_type in ['completed', 'created', 'updated', 'deleted']):
                return False

        if not self.config["batch_updates"]:
            return False

        user_id = event.user_id
        current_time = time.time()

        # Check last update time for this user
        last_update = self._last_update_times.get(user_id, 0)
        time_since_last = current_time - last_update

        # Use different throttle times based on event type
        if event.event_type == "cache_invalidate":
            # Use shorter throttle for user actions (200ms)
            throttle_time = 0.2
        else:
            # Use longer throttle for background updates (300ms default)
            throttle_time = self.config["update_throttle_seconds"]

        if time_since_last < throttle_time:
            # Store the most recent update to batch process later
            self._pending_updates[user_id] = event
            if not self.config["quiet_mode"]:
                logger.debug(
                    f"Throttling update for user {user_id}: {event.event_type} (last update {time_since_last:.3f}s ago, throttle: {throttle_time}s)")
            return True

        # Update last update time
        self._last_update_times[user_id] = current_time

        # Process any pending updates immediately
        if user_id in self._pending_updates:
            if not self.config["quiet_mode"]:
                logger.debug(f"Processing pending update for user {user_id}")
            del self._pending_updates[user_id]

        return False

    def _is_duplicate_event(self, event: DashboardEvent) -> bool:
        """Check if this event is a duplicate within the deduplication window"""
        current_time = time.time()
        user_id = event.user_id

        # Create a more specific unique key for this event
        action = event.details.get('action', '')
        entity_id = (event.details.get('todo_id') or
                     event.details.get('task_id') or
                     event.details.get('habit_id') or
                     event.details.get('calendar_event_id', ''))

        # More specific event key with event_type, action and entity_id
        if entity_id:
            event_key = f"{event.event_type}:{action}:{entity_id}"
        else:
            # For system updates (no entity), use coarser time buckets to avoid duplicates
            # Round to 1-second buckets for non-entity events
            rounded_time = int(current_time)
            event_key = f"{event.event_type}:{action}:{rounded_time}"

        # Clean up old events (older than dedup window)
        if user_id in self._recent_events:
            self._recent_events[user_id] = {
                key: timestamp for key, timestamp in self._recent_events[user_id].items()
                if current_time - timestamp < self._dedup_window
            }

        # Check if this event is a duplicate
        if user_id in self._recent_events and event_key in self._recent_events[user_id]:
            last_time = self._recent_events[user_id][event_key]
            if current_time - last_time < self._dedup_window:
                if not self.config["quiet_mode"]:
                    logger.debug(
                        f"Blocked duplicate event for user {user_id}: {event_key} (within {current_time - last_time:.3f}s)")
                return True

        # Record this event
        if user_id not in self._recent_events:
            self._recent_events[user_id] = {}
        self._recent_events[user_id][event_key] = current_time

        if not self.config["quiet_mode"]:
            logger.debug(
                f"Allowed event for user {user_id}: {event_key}")
        return False

    async def _notify_subscribers(self, event: DashboardEvent):
        # Check if real-time updates are disabled
        if not self.config["enable_realtime_updates"]:
            return

        # Check for duplicate events
        if self._is_duplicate_event(event):
            if not self.config["quiet_mode"]:
                logger.debug(
                    f"Skipping duplicate event for user {event.user_id}: {event.event_type}")
            return

        # Check if we should throttle this update
        if await self._should_throttle_update(event):
            if not self.config["quiet_mode"]:
                logger.debug(
                    f"Throttling update for user {event.user_id}, event: {event.event_type}")
            return

        # Publish to Redis channel for other services
        channel = f"dashboard_updates:{event.user_id}"
        await redis_client.publish(channel, json.dumps(event.to_dict()))

        # Broadcast to WebSocket clients if WebSocket manager is available
        if self.ws_manager:
            # For cache invalidation events, send a consolidated update message
            if event.event_type == "cache_invalidate":
                try:
                    # Send a single consolidated message that tells client to refresh
                    message = {
                        "type": "dashboard_update",
                        "timestamp": datetime.utcnow().isoformat(),
                        "data": {
                            "action": event.details.get("action", "data_changed"),
                            "entity_type": self._extract_entity_type(event.details),
                            "user_id": event.user_id
                        },
                        "requires_refresh": True
                    }

                    # Ensure we have WebSocket connections before trying to broadcast
                    if hasattr(self.ws_manager, 'active_connections') and event.user_id in self.ws_manager.active_connections:
                        await self.ws_manager.broadcast_to_user(event.user_id, message)
                        if not self.config["quiet_mode"]:
                            logger.info(
                                f"Sent consolidated dashboard update to user {event.user_id}")
                    return
                except Exception as e:
                    logger.error(
                        f"Error sending consolidated update: {e}")

            # Default case for other event types
            message = {
                "type": event.event_type,
                "timestamp": datetime.utcnow().isoformat(),
                "data": event.details
            }

            # Ensure we have WebSocket connections before trying to broadcast
            if hasattr(self.ws_manager, 'active_connections') and event.user_id in self.ws_manager.active_connections:
                await self.ws_manager.broadcast_to_user(event.user_id, message)
                if not self.config["quiet_mode"]:
                    logger.debug(
                        f"Broadcasted event to WebSocket clients for user {event.user_id}")

    def _extract_entity_type(self, details: dict) -> str:
        """Extract entity type from event details"""
        if "todo_id" in details:
            return "todo"
        elif "task_id" in details:
            return "task"
        elif "habit_id" in details:
            return "habit"
        elif "calendar_event_id" in details:
            return "calendar"
        else:
            return "unknown"

    async def start_go_metrics_subscriber(self):
        """Start listening for dashboard events from Go backend"""
        if self.is_subscribed:
            return

        self.is_subscribed = True
        logger.info("Starting Go backend dashboard metrics subscriber")

        # Subscribe to the dashboard events channel
        # The Go backend uses 'dashboard:events' as the channel name
        self.subscriber_task = asyncio.create_task(
            redis_pubsub_client.subscribe(
                "dashboard:events", self._handle_go_event)
        )

        # for currently connected users to avoid excessive subscriptions
        logger.info(
            "Starting individual user dashboard update channels subscriber")
        # Create task but don't await it - it will run in the background
        asyncio.create_task(
            self._subscribe_to_user_updates()
        )

    async def _subscribe_to_user_updates(self):
        """Helper method to subscribe to user-specific update channels"""
        try:
            await redis_pubsub_client.subscribe(
                "dashboard_updates:*", self._handle_go_event
            )
        except Exception as e:
            logger.error(f"Error subscribing to user update channels: {e}")

    async def _handle_go_event(self, event):
        """Handle dashboard events from Go backend"""
        try:
            if not self.config["quiet_mode"]:
                logger.debug(f"Received Go backend event: {event}")

            # Extract user ID from the event
            # Go backend sends a DashboardEvent with user_id as UUID
            if isinstance(event, dict) and "user_id" in event:
                # Convert UUID to string if needed
                user_id = str(event["user_id"])
                event_type = event.get("event_type", "unknown")

                logger.info(
                    f"Processing Go backend event: {event_type} for user {user_id}")

                # Invalidate cache for this user (both Redis and memory)
                cache_key = f"dashboard:metrics:{user_id}"
                await redis_client.delete(cache_key)

                # Also clear memory cache to prevent stale data
                if user_id in self._memory_cache:
                    del self._memory_cache[user_id]
                    logger.debug(f"Cleared memory cache for user {user_id}")

                logger.info(
                    f"Invalidated dashboard cache for user {user_id} due to Go backend event: {event_type}")

                # Create a Python-style dashboard event and notify subscribers
                dashboard_event = DashboardEvent(
                    event_type=event_type,
                    user_id=user_id,
                    entity_id=event.get("entity_id", ""),
                    details=event.get("details", {})
                )

                # Add deduplication by event type
                # Skip sending metrics_update events too frequently
                if event_type == "metrics_update":
                    # Check if we've sent a similar event recently
                    current_time = time.time()
                    last_update = self._last_update_times.get(
                        f"metrics:{user_id}", 0)

                    # Only allow metrics updates every 2 seconds to avoid spam
                    if current_time - last_update < 2.0:
                        logger.debug(
                            f"Skipping frequent metrics_update for user {user_id}")
                        return

                    # Update timestamp for this metrics update
                    self._last_update_times[f"metrics:{user_id}"] = current_time

                await self._notify_subscribers(dashboard_event)
            else:
                logger.warning(f"Received malformed Go backend event: {event}")
        except Exception as e:
            logger.error(
                f"Error handling Go backend event: {e}", exc_info=True)

    async def _handle_notes_event(self, event):
        """Handle events from Notes server"""
        try:
            logger.debug(f"Received Notes server event: {event}")

            if isinstance(event, dict) and "user_id" in event:
                user_id = str(event["user_id"])
                event_type = event.get("event_type", "unknown")
                details = event.get("details", {})

                logger.info(
                    f"Processing Notes server event: {event_type} for user {user_id}")

                # Invalidate cache for this user
                cache_key = f"dashboard:metrics:{user_id}"
                await self.redis_client.delete(cache_key)
                logger.info(
                    f"Invalidated dashboard cache for user {user_id} due to Notes server event: {event_type}")

                # If this is a metrics update event, we could potentially fetch new metrics immediately
                if event_type == events.DashboardEventMetricsUpdate:
                    logger.info(
                        f"Metrics update event received for user {user_id}")

                # Create a Python-style dashboard event and notify subscribers
                dashboard_event = DashboardEvent(
                    event_type=event_type,
                    user_id=user_id,
                    entity_id=event.get("entity_id", ""),
                    details=details
                )
                await self._notify_subscribers(dashboard_event)
            else:
                logger.warning(
                    f"Received malformed Notes server event: {event}")
        except Exception as e:
            logger.error(
                f"Error handling Notes server event: {e}", exc_info=True)

    async def start_notes_metrics_subscriber(self):
        """Start subscriber for Notes server events"""
        if self.is_subscribed:
            return

        self.is_subscribed = True
        logger.info("Starting Notes server dashboard metrics subscriber")

        # Subscribe to the dashboard events channel
        self.subscriber_task = asyncio.create_task(
            redis_pubsub_client.subscribe(
                "dashboard:events", self._handle_notes_event)
        )

    def get_metrics_statistics(self):
        """Get statistics about dashboard metrics for monitoring"""
        stats = {
            "cache_hit_rate": 0,
            "memory_cache_hit_rate": 0,
            "avg_fetch_time_ms": 0,
            "error_rate": 0,
            "memory_cache_size": len(self._memory_cache),
            "circuit_breaker_status": {
                "go_backend": {
                    "state": "open" if self._go_backend_circuit.is_open else "closed",
                    "failure_count": self._go_backend_circuit.failure_count,
                    "last_failure": self._go_backend_circuit.last_failure_time
                },
                "notes_server": {
                    "state": "open" if self._notes_server_circuit.is_open else "closed",
                    "failure_count": self._notes_server_circuit.failure_count,
                    "last_failure": self._notes_server_circuit.last_failure_time
                }
            }
        }

        # Calculate cache hit rates
        total_redis_requests = self._metrics["cache_hits"] + \
            self._metrics["cache_misses"]
        if total_redis_requests > 0:
            stats["cache_hit_rate"] = self._metrics["cache_hits"] / \
                total_redis_requests

        total_memory_requests = self._metrics["memory_cache_hits"] + \
            self._metrics["memory_cache_misses"]
        if total_memory_requests > 0:
            stats["memory_cache_hit_rate"] = self._metrics["memory_cache_hits"] / \
                total_memory_requests

        # Calculate average fetch time
        if self._metrics["fetch_times"]:
            stats["avg_fetch_time_ms"] = sum(
                self._metrics["fetch_times"]) / len(self._metrics["fetch_times"])

        # Calculate error rate (errors per 100 requests)
        total_requests = total_redis_requests
        if total_requests > 0:
            stats["error_rate"] = (
                self._metrics["errors"] / total_requests) * 100

        return stats

    async def close(self):
        """Cleanup resources"""
        if self.is_subscribed:
            await self.pubsub_manager.unsubscribe()
            self.is_subscribed = False
            logger.info("Unsubscribed from Notes server dashboard events")

        if self.subscriber_task:
            self.subscriber_task.cancel()
            try:
                await self.subscriber_task
            except asyncio.CancelledError:
                pass
            self.subscriber_task = None

        if self.session:
            await self.session.close()
            self.session = None


dashboard_cache = DashboardCache()
