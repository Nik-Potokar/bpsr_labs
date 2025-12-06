#!/bin/bash

echo "========================================="
echo "BPSR API Market Endpoints Test"
echo "========================================="
echo ""

API_URL="http://localhost:8989"

echo "1. Testing Item Catalog (should work immediately)..."
echo "GET $API_URL/api/items/10002"
curl -s "$API_URL/api/items/10002" | python3 -m json.tool 2>/dev/null || echo "Error: API not responding or invalid JSON"
echo ""
echo ""

echo "2. Testing Item Search..."
echo "GET $API_URL/api/items/search?q=Luno"
curl -s "$API_URL/api/items/search?q=Luno" | python3 -m json.tool 2>/dev/null || echo "Error: API not responding"
echo ""
echo ""

echo "3. Testing Market Listings (needs market_listings.json)..."
echo "GET $API_URL/api/market/listings"
curl -s "$API_URL/api/market/listings" | python3 -m json.tool 2>/dev/null || echo "Error: API not responding"
echo ""
echo ""

echo "4. Testing Scene Data (original endpoint)..."
echo "GET $API_URL/api/scene"
curl -s "$API_URL/api/scene" | python3 -m json.tool 2>/dev/null || echo "Error: API not responding"
echo ""
echo ""

echo "========================================="
echo "Test Complete"
echo "========================================="
