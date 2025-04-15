import asyncio
import redis.asyncio as redis


async def main():
    r = redis.from_url("redis://localhost:6380/1")  
    pubsub = r.pubsub()
    user_id = "11e41910-a77c-4818-b073-28019b0fbc92"
    channel = f"dashboard_updates:{user_id}"
    await pubsub.subscribe(channel)
    print(f"Subscribed to {channel}")
    async for message in pubsub.listen():
        if message['type'] == 'message':
            print("Received:", message['data'])

asyncio.run(main())
