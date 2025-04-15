#!/usr/bin/env python3
"""
Test script to verify WebSocket message deduplication is working correctly
"""

import asyncio
import websockets
import json
import time
from datetime import datetime


async def test_websocket_deduplication():
    """Test WebSocket connection and message deduplication"""

    # You'll need to replace this with a valid token
    token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTFlNDE5MTAtYTc3Yy00ODE4LWIwNzMtMjgwMTliMGZiYzkyIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnVwZGF0ZSIsInRhc2tzOmNyZWF0ZSIsInRhc2tzOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwib3JnYW5pemF0aW9uczpyZWFkIl0sImV4cCI6MTc0OTUyNjg2NSwibmJmIjoxNzQ5NDQwNDY1LCJpYXQiOjE3NDk0NDA0NjV9.QQ49tm4Vagf0tz3__Hb7m4K0gk3BP_d7YFTGRWppbfE"

    uri = f"ws://localhost:8001/ws/dashboard?token={token}"

    print(f"Connecting to WebSocket: {uri}")

    try:
        async with websockets.connect(uri) as websocket:
            print("Connected to WebSocket")

            # Send ping to test connection
            await websocket.send(json.dumps({"type": "ping"}))
            print("Sent ping")

            # Listen for messages for 30 seconds
            start_time = time.time()
            message_count = 0
            dashboard_updates = 0
            cache_invalidates = 0

            while time.time() - start_time < 30:
                try:
                    message = await asyncio.wait_for(websocket.recv(), timeout=1.0)
                    data = json.loads(message)
                    message_count += 1

                    print(
                        f"[{message_count}] {datetime.now().strftime('%H:%M:%S.%f')[:-3]} - Type: {data.get('type', 'unknown')}")

                    if data.get("type") == "dashboard_update":
                        dashboard_updates += 1
                        print(
                            f"   └─ Dashboard update #{dashboard_updates}: {data.get('data', {})}")
                    elif data.get("type") == "cache_invalidate":
                        cache_invalidates += 1
                        print(
                            f"   └─ Cache invalidate #{cache_invalidates}: {data.get('data', {})}")
                    elif data.get("type") == "initial_metrics":
                        print(f"   └─ Initial metrics received")
                    elif data.get("type") == "fresh_metrics":
                        print(f"   └─ Fresh metrics received")
                    elif data.get("type") in ["pong", "connected"]:
                        print(f"   └─ Connection message")
                    else:
                        print(f"   └─ Other: {json.dumps(data, indent=2)}")

                except asyncio.TimeoutError:
                    continue
                except websockets.exceptions.ConnectionClosed:
                    print("WebSocket connection closed")
                    break

            print(f"\nTest Results (30 seconds):")
            print(f"   • Total messages: {message_count}")
            print(f"   • Dashboard updates: {dashboard_updates}")
            print(f"   • Cache invalidations: {cache_invalidates}")
            print(f"   • Messages per second: {message_count / 30:.2f}")

            if cache_invalidates > dashboard_updates * 2:
                print(
                    "Warning: Too many cache invalidation messages (potential duplication)")
            else:
                print("Message frequency looks reasonable")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    print("Testing WebSocket message deduplication")
    print("=" * 50)

    # Instructions for getting a token
    print("To run this test:")
    print("1. Get a valid JWT token from your browser's localStorage")
    print("2. Replace 'your_jwt_token_here' with your actual token")
    print("3. Make sure your backend is running on localhost:8001")
    print("4. Run some actions (complete todos, etc.) while the test runs")
    print()

    # Run the test
    asyncio.run(test_websocket_deduplication())
