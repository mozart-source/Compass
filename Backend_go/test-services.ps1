Write-Host "🔍 Testing COMPASS services..." -ForegroundColor Cyan
Write-Host ""

# Test nginx directly
Write-Host "📋 Testing Nginx..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:8081/nginx-test"
Write-Host ""

# Test static file serving
Write-Host "📋 Testing static file serving..." -ForegroundColor Yellow
$response = Invoke-WebRequest -Uri "http://localhost:8081/test.html" -Method Head
Write-Host $response.StatusCode $response.StatusDescription
Write-Host ""

# Test Go backend health
Write-Host "📋 Testing Go backend health..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/health"
    $health | ConvertTo-Json
} catch {
    Write-Host "Failed to get Go backend health: $_" -ForegroundColor Red
}
Write-Host ""

# Test Python backend health
Write-Host "📋 Testing Python backend health..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/health"
    $health | ConvertTo-Json
} catch {
    Write-Host "Failed to get Python backend health: $_" -ForegroundColor Red
}
Write-Host ""

# Test Notes server health
Write-Host "📋 Testing Notes server health..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/notes/health"
    $health | ConvertTo-Json
} catch {
    Write-Host "Failed to get Notes server health: $_" -ForegroundColor Red
}
Write-Host ""

# Test frontend
Write-Host "📋 Testing frontend..." -ForegroundColor Yellow
$response = Invoke-WebRequest -Uri "http://localhost:8081/" -Method Head
Write-Host $response.StatusCode $response.StatusDescription
Write-Host ""

Write-Host "✅ Tests completed!" -ForegroundColor Green 