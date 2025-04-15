import asyncio
import redis.asyncio as redis
import json


async def main():
    r = redis.from_url("redis://localhost:6380/1")
    user_id = "11e41910-a77c-4818-b073-28019b0fbc92"
    channel = f"dashboard_updates:{user_id}"
    payload = {"event": "test_event", "data": {"foo": "bar"}}
    await r.publish(channel, json.dumps(payload))
    print("Published test event.")

asyncio.run(main())
