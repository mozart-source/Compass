#!/usr/bin/env python
import asyncio
import websockets
import json
import sys
import requests
import time

# Replace with a valid JWT token
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDFlMTVjMzEtNWRkYS00M2Q4LWFlYTUtMmQ2MDQ5OTM2YzFkIiwiZW1haWwiOiJhQGEuY29tIiwicm9sZXMiOlsidXNlciJdLCJvcmdfaWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJwZXJtaXNzaW9ucyI6WyJ0YXNrczpyZWFkIiwidGFza3M6dXBkYXRlIiwidGFza3M6Y3JlYXRlIiwicHJvamVjdHM6cmVhZCIsIm9yZ2FuaXphdGlvbnM6cmVhZCJdLCJleHAiOjE3NDk5OTE3ODcsIm5iZiI6MTc0OTkwNTM4NywiaWF0IjoxNzQ5OTA1Mzg3fQ.9CxvOOi6giwVX0uYowaExP4rwqiKwAOSeBEuCzJQs9I"


def create_report():
    """Create a new report using the REST API."""
    url = "http://localhost:8001/api/v1/reports"
    headers = {
        "Authorization": f"Bearer {TOKEN}",
        "Content-Type": "application/json"
    }
    payload = {
        "title": "Test Report for WebSocket",
        "type": "activity",
        "time_range": {
            "start_date": "2025-06-01",
            "end_date": "2025-06-13"
        }
    }

    print("Creating a new report...")
    response = requests.post(url, headers=headers, json=payload)

    if response.status_code == 200:
        data = response.json()
        report_id = data.get("report_id")
        print(f"Report created successfully with ID: {report_id}")
        return report_id
    else:
        print(
            f"Failed to create report: {response.status_code} - {response.text}")
        return None


async def test_websocket_connection(report_id):
    """Test the WebSocket connection for real-time report updates."""
    # Test all possible WebSocket URLs
    urls = [
        f"ws://localhost:8001/api/v1/ws/reports/{report_id}?token={TOKEN}",
        f"ws://localhost:8001/api/v1/ws/reports/{report_id}/generate?token={TOKEN}",
        f"ws://localhost:8001/ws/reports/{report_id}?token={TOKEN}",
        f"ws://localhost:8001/ws/reports/{report_id}/generate?token={TOKEN}"
    ]

    # Also try with Authorization header instead of query param
    headers = {"Authorization": f"Bearer {TOKEN}"}

    for url in urls:
        print(f"Testing connection to: {url}")
        try:
            # Try with token in URL
            async with websockets.connect(url) as websocket:
                print("Connected successfully!")

                # Wait for the initial connection message
                response = await websocket.recv()
                print(f"Received: {response}")

                # Send a ping command
                await websocket.send(json.dumps({"action": "ping"}))
                response = await websocket.recv()
                print(f"Ping response: {response}")

                # Send a generate command
                print("Sending generate command...")
                await websocket.send(json.dumps({"action": "generate"}))

                # Wait for a few responses
                for _ in range(3):
                    try:
                        response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                        print(f"Generate response: {response}")
                    except asyncio.TimeoutError:
                        print("Timed out waiting for response")
                        break

                # Close the connection
                await websocket.close()
                print("Connection closed successfully")
                return True
        except Exception as e:
            print(f"Error connecting to {url}: {e}")

            # If URL doesn't have token, try with Authorization header
            if "?token=" not in url:
                try:
                    print(f"Retrying with Authorization header")
                    async with websockets.connect(url, extra_headers=headers) as websocket:
                        print("Connected successfully with Authorization header!")
                        # Rest of the code would be the same...
                        return True
                except Exception as e2:
                    print(f"Error connecting with Authorization header: {e2}")

    print("Failed to connect to any WebSocket URL")
    return False


async def main():
    # Step 1: Create a report
    report_id = create_report()
    if not report_id:
        print("Failed to create report, exiting.")
        return False

    # Give the server a moment to process the report
    print("Waiting for the server to process the report...")
    time.sleep(2)

    # Step 2: Test WebSocket connection with the new report ID
    return await test_websocket_connection(report_id)

if __name__ == "__main__":
    success = asyncio.run(main())
    sys.exit(0 if success else 1)
