from typing import Callable, Dict, List
from Backend.events.event_registry import EVENT_TYPES
import logging

logger = logging.getLogger(__name__)


class EventDispatcher:
    def __init__(self):
        self.listeners: Dict[str, List[Callable]] = {
            event: [] for event in EVENT_TYPES}

    def register_listener(self, event_type: str, callback: Callable):
        if event_type in self.listeners:
            self.listeners[event_type].append(callback)
        else:
            raise ValueError(f"Unknown event type: {event_type}")

    async def dispatch(self, event_type: str, payload: dict):
        if event_type in self.listeners:
            for listener in self.listeners[event_type]:
                try:
                    await listener(payload)
                except Exception as e:
                    logger.error(f"Error processing {event_type}: {e}")

# Create a singleton instance
dispatcher = EventDispatcher()
