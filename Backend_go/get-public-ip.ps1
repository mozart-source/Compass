Write-Host "🌐 Getting your public IP address for MongoDB Atlas whitelisting..." -ForegroundColor Cyan
Write-Host ""

try {
    # Method 1: Using Invoke-RestMethod with ipinfo.io
    Write-Host "📍 Method 1 - ipinfo.io:" -ForegroundColor Yellow
    $PUBLIC_IP1 = (Invoke-RestMethod -Uri "https://ipinfo.io/ip").Trim()
    Write-Host "   Your public IP: $PUBLIC_IP1" -ForegroundColor Green
    
    # Method 2: Using Invoke-RestMethod with ifconfig.me
    Write-Host "📍 Method 2 - ifconfig.me:" -ForegroundColor Yellow
    $PUBLIC_IP2 = (Invoke-RestMethod -Uri "https://ifconfig.me").Trim()
    Write-Host "   Your public IP: $PUBLIC_IP2" -ForegroundColor Green
    
    # Method 3: Using Invoke-RestMethod with icanhazip.com
    Write-Host "📍 Method 3 - icanhazip.com:" -ForegroundColor Yellow
    $PUBLIC_IP3 = (Invoke-RestMethod -Uri "https://icanhazip.com").Trim()
    Write-Host "   Your public IP: $PUBLIC_IP3" -ForegroundColor Green
}
catch {
    Write-Host "❌ Error getting IP: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "💡 Try manually visiting: https://whatismyipaddress.com/" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "MongoDB Atlas Whitelisting Instructions:" -ForegroundColor Magenta
Write-Host "   1. Go to MongoDB Atlas -> Network Access" -ForegroundColor White
Write-Host "   2. Click Add IP Address" -ForegroundColor White
Write-Host "   3. Add: $PUBLIC_IP1" -ForegroundColor Green
Write-Host "   4. Or for development, use: 0.0.0.0/0 (allows all IPs - not recommended for production)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Note: Your IP may change if you have a dynamic IP from your ISP" -ForegroundColor Red
Write-Host "In that case, you might need to update the whitelist periodically" -ForegroundColor Red 