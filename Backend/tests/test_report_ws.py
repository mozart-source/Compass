#!/usr/bin/env python
import asyncio
import websockets
import json
import sys


async def test_report_websocket():
    """Test the report WebSocket connection."""

    # Replace with an actual report ID from your database
    report_id = "684c409a16f4255ab670f573"

    # Replace with a valid JWT token
    token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTFlNDE5MTAtYTc3Yy00ODE4LWIwNzMtMjgwMTliMGZiYzkyIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnVwZGF0ZSIsInRhc2tzOmNyZWF0ZSIsInRhc2tzOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwib3JnYW5pemF0aW9uczpyZWFkIl0sImV4cCI6MTc0OTkxMTkxMywibmJmIjoxNzQ5ODI1NTEzLCJpYXQiOjE3NDk4MjU1MTN9.hEcQ3G0BI71_O9ddWCCSm-xGlpYexXSAKb1iq_Suw9g"

    # Test both possible WebSocket URLs
    urls = [
        f"ws://localhost:8001/api/v1/ws/reports/{report_id}?token={token}",
        f"ws://localhost:8001/ws/reports/{report_id}?token={token}"
    ]

    for url in urls:
        print(f"Testing connection to: {url}")
        try:
            async with websockets.connect(url) as websocket:
                print("Connected successfully!")

                # Wait for the initial connection message
                response = await websocket.recv()
                print(f"Received: {response}")

                # Send a ping command
                await websocket.send(json.dumps({"action": "ping"}))
                response = await websocket.recv()
                print(f"Ping response: {response}")

                # Close the connection
                await websocket.close()
                print("Connection closed successfully")
                return True
        except Exception as e:
            print(f"Error connecting to {url}: {e}")

    print("Failed to connect to any WebSocket URL")
    return False

if __name__ == "__main__":
    success = asyncio.run(test_report_websocket())
    sys.exit(0 if success else 1)
