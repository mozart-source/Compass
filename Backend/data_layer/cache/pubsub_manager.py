import asyncio
import json
from data_layer.cache.redis_client import redis_client, redis_pubsub_client
from fastapi.encoders import jsonable_encoder


class PubSubManager:
    def __init__(self):
        self.listeners = []

    async def subscribe(self, callback):
        self.listeners.append(callback)
        await redis_pubsub_client.subscribe(
            "dashboard_events", callback)

    async def unsubscribe(self):
        self.listeners = []
        await redis_pubsub_client.close()

    async def notify(self, event):
        for cb in self.listeners:
            await cb(event)

    async def publish(self, user_id, event, data):
        channel = f"dashboard_updates:{user_id}"
        serializable_data = jsonable_encoder(data)
        await redis_client.publish(channel, json.dumps({"event": event, "data": serializable_data}))


pubsub_manager = PubSubManager()
