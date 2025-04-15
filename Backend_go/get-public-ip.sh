#!/bin/bash

echo "🌐 Getting your public IP address for MongoDB Atlas whitelisting..."
echo ""

# Method 1: Using curl with ipinfo.io
echo "📍 Method 1 - ipinfo.io:"
PUBLIC_IP1=$(curl -s https://ipinfo.io/ip)
echo "   Your public IP: $PUBLIC_IP1"

# Method 2: Using curl with ifconfig.me  
echo "📍 Method 2 - ifconfig.me:"
PUBLIC_IP2=$(curl -s https://ifconfig.me)
echo "   Your public IP: $PUBLIC_IP2"

# Method 3: Using curl with icanhazip.com
echo "📍 Method 3 - icanhazip.com:"
PUBLIC_IP3=$(curl -s https://icanhazip.com)
echo "   Your public IP: $PUBLIC_IP3"

echo ""
echo "🔒 MongoDB Atlas Whitelisting Instructions:"
echo "   1. Go to MongoDB Atlas → Network Access"
echo "   2. Click 'Add IP Address'"
echo "   3. Add: $PUBLIC_IP1"
echo "   4. Or for development, use: 0.0.0.0/0 (allows all IPs - not recommended for production)"
echo ""
echo "⚠️  Note: Your IP may change if you have a dynamic IP from your ISP"
echo "   In that case, you might need to update the whitelist periodically" 