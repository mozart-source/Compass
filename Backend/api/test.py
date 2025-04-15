import websockets
import asyncio
import json

async def test_ws():
    uri = "ws://localhost:8000/system-metrics/ws"
    headers = [("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTFlNDE5MTAtYTc3Yy00ODE4LWIwNzMtMjgwMTliMGZiYzkyIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnVwZGF0ZSIsInRhc2tzOmNyZWF0ZSIsInRhc2tzOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwib3JnYW5pemF0aW9uczpyZWFkIl0sImV4cCI6MTc0OTAyOTg2OCwibmJmIjoxNzQ4OTQzNDY4LCJpYXQiOjE3NDg5NDM0Njh9.cLMTyegnisdEh-k3jEbuFQv0RqMaoIsnCLVjGW78CyQ")]
    async with websockets.connect(uri, extra_headers=headers) as ws:
        print("Connected!")

asyncio.run(test_ws())
