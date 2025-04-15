#!/bin/bash

echo "🔍 Testing COMPASS services..."
echo ""

# Test nginx directly
echo "📋 Testing Nginx..."
curl -s http://localhost:8081/nginx-test
echo ""

# Test static file serving
echo "📋 Testing static file serving..."
curl -s -I http://localhost:8081/test.html | head -1
echo ""

# Test Go backend health
echo "📋 Testing Go backend health..."
curl -s http://localhost:8081/health | jq || echo "Failed to get Go backend health"
echo ""

# Test Python backend health
echo "📋 Testing Python backend health..."
curl -s http://localhost:8081/api/v1/health | jq || echo "Failed to get Python backend health"
echo ""

# Test Notes server health
echo "📋 Testing Notes server health..."
curl -s http://localhost:8081/notes/health | jq || echo "Failed to get Notes server health"
echo ""

# Test frontend
echo "📋 Testing frontend..."
curl -s -I http://localhost:8081/ | head -1
echo ""

echo "✅ Tests completed!" 