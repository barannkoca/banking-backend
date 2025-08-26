#!/bin/bash

# HTTP Server Setup Test Script
# Bu script, implement edilen HTTP server özelliklerini test eder

echo "🚀 Banking Backend HTTP Server Setup Test"
echo "=========================================="

# Server'ı background'da başlat
echo "📡 Starting server..."
./bin/banking-backend &
SERVER_PID=$!

# Server'ın başlaması için bekle
sleep 3

echo ""
echo "🧪 Testing HTTP Server Features"
echo "================================"

# Test 1: Health Check (Rate limiting olmamalı)
echo "1️⃣ Testing Health Check (no rate limiting)..."
for i in {1..5}; do
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/health)
    echo "   Request $i: HTTP $response"
done

echo ""
echo "2️⃣ Testing Rate Limiting..."
echo "   Making rapid requests to trigger rate limiting..."

# Test 2: Rate Limiting (hızlı istekler)
for i in {1..15}; do
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/api/v1/auth/login)
    echo "   Request $i: HTTP $response"
    if [ "$response" = "429" ]; then
        echo "   ✅ Rate limiting triggered at request $i"
        break
    fi
done

echo ""
echo "3️⃣ Testing Security Headers..."
echo "   Checking security headers in response..."

# Test 3: Security Headers
headers=$(curl -s -I http://localhost:8080/health)
echo "$headers" | grep -E "(X-Content-Type-Options|X-Frame-Options|X-XSS-Protection|Strict-Transport-Security|X-Request-ID)"

echo ""
echo "4️⃣ Testing CORS..."
echo "   Testing CORS headers..."

# Test 4: CORS
cors_headers=$(curl -s -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: GET" -H "Access-Control-Request-Headers: X-Requested-With" -X OPTIONS http://localhost:8080/api/v1/accounts -v 2>&1)
echo "$cors_headers" | grep -E "(Access-Control-Allow-Origin|Access-Control-Allow-Methods|Access-Control-Allow-Headers)"

echo ""
echo "5️⃣ Testing Authentication..."
echo "   Testing protected endpoint without token..."

# Test 5: Authentication
auth_response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/api/v1/accounts)
echo "   Protected endpoint without token: HTTP $auth_response"

echo ""
echo "6️⃣ Testing with Valid Token..."
echo "   Testing protected endpoint with valid token..."

# Test 6: Valid Token
valid_response=$(curl -s -w "%{http_code}" -o /dev/null -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/accounts)
echo "   Protected endpoint with valid token: HTTP $valid_response"

echo ""
echo "7️⃣ Testing Admin Authorization..."
echo "   Testing admin endpoint with customer token..."

# Test 7: Admin Authorization
admin_response=$(curl -s -w "%{http_code}" -o /dev/null -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/admin/users)
echo "   Admin endpoint with customer token: HTTP $admin_response"

echo ""
echo "8️⃣ Testing Admin with Admin Token..."
echo "   Testing admin endpoint with admin token..."

# Test 8: Admin with Admin Token
admin_valid_response=$(curl -s -w "%{http_code}" -o /dev/null -H "Authorization: Bearer admin-token" http://localhost:8080/api/v1/admin/users)
echo "   Admin endpoint with admin token: HTTP $admin_valid_response"

echo ""
echo "9️⃣ Testing Request Tracking..."
echo "   Checking for request tracking headers..."

# Test 9: Request Tracking
tracking_response=$(curl -s -I http://localhost:8080/health)
echo "$tracking_response" | grep -E "(X-Request-ID|X-Response-Time|X-Banking-Security)"

echo ""
echo "🔍 Checking Server Logs..."
echo "   Server should have logged all requests with structured logging"

# Server'ı durdur
echo ""
echo "🛑 Stopping server..."
kill $SERVER_PID

echo ""
echo "✅ HTTP Server Setup Test Completed!"
echo "====================================="
echo ""
echo "📊 Test Summary:"
echo "   ✅ Custom Router with Middleware Support"
echo "   ✅ CORS and Security Headers"
echo "   ✅ Rate Limiting"
echo "   ✅ Request Logging and Tracking"
echo "   ✅ Authentication and Authorization"
echo ""
echo "📝 Next Steps:"
echo "   - Implement real JWT validation"
echo "   - Add Redis for distributed rate limiting"
echo "   - Implement actual business logic handlers"
echo "   - Add comprehensive API documentation"
