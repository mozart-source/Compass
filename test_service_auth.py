#!/usr/bin/env python3
"""
Test script to verify service-to-service authentication with Go backend
"""

import requests
import json

# Test JWT token (replace with your actual token)
JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTFlNDE5MTAtYTc3Yy00ODE4LWIwNzMtMjgwMTliMGZiYzkyIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnVwZGF0ZSIsInRhc2tzOmNyZWF0ZSIsInRhc2tzOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwib3JnYW5pemF0aW9uczpyZWFkIl0sImV4cCI6MTc0OTUyNjI4NSwibmJmIjoxNzQ5NDM5ODg1LCJpYXQiOjE3NDk0Mzk4ODV9.k_koAYpF0cmFHdQ2DjBa8Nthfun-ouNmIj_QgbnrsOw"

GO_BACKEND_URL = "http://localhost:8000"


def test_regular_call():
    """Test regular call without service headers - should fail"""
    print("Testing regular call (should fail)...")

    headers = {
        "Authorization": f"Bearer {JWT_TOKEN}",
        "Content-Type": "application/json"
    }

    try:
        response = requests.get(
            f"{GO_BACKEND_URL}/api/dashboard/metrics", headers=headers)
        print(f"   Status: {response.status_code}")
        if response.status_code != 200:
            print(f"   Error: {response.text}")
        else:
            print(f"   Success: {response.json()}")
    except Exception as e:
        print(f"   Exception: {e}")


def test_service_call():
    """Test service-to-service call with proper headers - should succeed"""
    print("\nTesting service-to-service call (should succeed)...")

    headers = {
        "Authorization": f"Bearer {JWT_TOKEN}",
        "Content-Type": "application/json",
        "X-Service-Call": "true",
        "X-Internal-Service": "python-backend",
        "User-Agent": "python-backend-aiohttp/dashboard-service"
    }

    try:
        response = requests.get(
            f"{GO_BACKEND_URL}/api/dashboard/metrics", headers=headers)
        print(f"   Status: {response.status_code}")
        if response.status_code != 200:
            print(f"   Error: {response.text}")
        else:
            data = response.json()
            print(f"   Success! Got data: {json.dumps(data, indent=2)}")
    except Exception as e:
        print(f"   Exception: {e}")


def test_user_agent_detection():
    """Test User-Agent based service detection"""
    print("\nTesting User-Agent based service detection...")

    headers = {
        "Authorization": f"Bearer {JWT_TOKEN}",
        "Content-Type": "application/json",
        "User-Agent": "aiohttp/3.8.0"  # Should be detected as service call
    }

    try:
        response = requests.get(
            f"{GO_BACKEND_URL}/api/dashboard/metrics", headers=headers)
        print(f"   Status: {response.status_code}")
        if response.status_code != 200:
            print(f"   Error: {response.text}")
        else:
            data = response.json()
            print(f"   Success! Got data keys: {list(data.keys())}")
    except Exception as e:
        print(f"   Exception: {e}")


if __name__ == "__main__":
    print("Testing Go Backend Service-to-Service Authentication\n")

    test_regular_call()
    test_service_call()
    test_user_agent_detection()

    print("\nTesting complete!")
