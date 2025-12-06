# Debugging "No Data Appears" Issue

## Quick Checklist

### ✅ Step 1: Is the API Running?

```bash
# Check if the API is running on port 8989
curl http://localhost:8989/api/items/10002
```

**Expected Response:**
```json
{"code":0,"msg":"OK","data":{"item_id":10002,"item_name":"Luno"}}
```

**If you get:**
- "Connection refused" → API is not running
- Empty/no response → Check if port 8989 is blocked
- JSON response → API is running! ✅

---

### ✅ Step 2: Check What the API Says on Startup

Look for these specific lines when you start the API:

```
Loaded 6367 items into catalog         ← Item catalog loaded ✅
Market data reloader started           ← Market feature active ✅
Imported 3 market listings             ← Data loaded! ✅
  OR
Initial market data load failed        ← File not found! ❌
Service started at: http://127.0.0.1:8989
```

**Copy the exact startup logs and share them!**

---

### ✅ Step 3: Verify File Location

The API looks for `market_listings.json` in the **current directory** when you run it.

```bash
# ARE YOU IN THE RIGHT DIRECTORY?
pwd   # Should show: .../BPSR-API-2.0-main

# CHECK IF FILE EXISTS HERE
ls -la market_listings.json

# If not found, create it or specify path:
./BPSR-API-2.0.exe --marketData /full/path/to/market_listings.json
```

---

### ✅ Step 4: Test Each Endpoint

**Test 1: Item Catalog (always works)**
```bash
curl http://localhost:8989/api/items/10002
```
Should return: `{"code":0,"msg":"OK","data":{"item_id":10002,"item_name":"Luno"}}`

**Test 2: Market Listings**
```bash
curl http://localhost:8989/api/market/listings
```
**If you get:**
```json
{"code":0,"msg":"OK","data":{"listings":[],"count":0}}
```
This means **API is working BUT no market data is loaded!**

**If you get:**
```json
{"code":0,"msg":"OK","data":{"listings":[...],"count":3}}
```
This means **it's working! You have 3 listings!**

---

### ✅ Step 5: Force Reload

```bash
# Stop the API (Ctrl+C)

# Make sure you're in the right directory
cd BPSR-API-2.0-main

# Verify file exists
ls -la market_listings.json

# Start with explicit path
./BPSR-API-2.0.exe --network auto --marketData market_listings.json

# Watch for "Imported X market listings" in the logs!
```

---

## Common Problems & Solutions

### Problem 1: "Connection refused"
**Cause:** API is not running
**Solution:**
```bash
cd BPSR-API-2.0-main
./BPSR-API-2.0.exe --network auto
```

### Problem 2: `{"listings":[],"count":0}`
**Cause:** API is running but no market data loaded
**Check:**
1. Is `market_listings.json` in the same directory as the .exe?
2. Did you see "Imported X listings" in startup logs?
3. Try restarting the API

**Solution:**
```bash
# Verify file is in right place
cd BPSR-API-2.0-main
ls -la market_listings.json  # Must see the file!

# Restart API
./BPSR-API-2.0.exe --network auto --marketData ./market_listings.json
```

### Problem 3: File exists but API says "file not found"
**Cause:** Wrong working directory
**Solution:**
```bash
# Always run from BPSR-API-2.0-main directory!
cd /full/path/to/BPSR-API-2.0-main
./BPSR-API-2.0.exe --network auto

# OR use absolute path:
./BPSR-API-2.0.exe --network auto --marketData "C:\full\path\to\market_listings.json"
```

### Problem 4: Invalid JSON format
**Cause:** market_listings.json has wrong format
**Solution:** Use the example file I created, or validate JSON:
```bash
# Validate JSON (if you have Python)
python -m json.tool market_listings.json

# Should show formatted JSON without errors
```

---

## What to Share for Help

If still not working, share these:

1. **Exact API startup logs** (first 10 lines)
2. **Current directory**: Output of `pwd`
3. **File check**: Output of `ls -la market_listings.json`
4. **API response**: Output of `curl http://localhost:8989/api/market/listings`
5. **Item test**: Output of `curl http://localhost:8989/api/items/10002`

---

## Quick Test - Copy & Paste This

```bash
# === DIAGNOSTIC TEST ===
echo "1. Current directory:"
pwd

echo ""
echo "2. Market data file:"
ls -la market_listings.json

echo ""
echo "3. Testing API - Item catalog:"
curl -s http://localhost:8989/api/items/10002

echo ""
echo "4. Testing API - Market listings:"
curl -s http://localhost:8989/api/market/listings

echo ""
echo "=== TEST COMPLETE ==="
```

Copy the output and share it!

---

## Expected Working Output

```bash
1. Current directory:
/c/Users/Admin/Desktop/BPSR-API-2.0-main

2. Market data file:
-rw-r--r-- 1 Admin 1377 Dec 6 market_listings.json

3. Testing API - Item catalog:
{"code":0,"msg":"OK","data":{"item_id":10002,"item_name":"Luno"}}

4. Testing API - Market listings:
{"code":0,"msg":"OK","data":{"count":3,"listings":[{"bind_flag":false,"captured_at":1733497200,"item_id":10002,"item_name":"Luno","listing_guid":"9d9dfe8c-55a2-4fb9-b903-1b63a4514c52","listed_at":1733493600,"price_luno":4500,"quantity":3,"unit_price":1500}...]}}
```

If you see this ↑ **IT'S WORKING!** You have 3 test listings loaded! 🎉

---

## Next Steps After It Works

1. **Keep the test data** or replace with real data:
   ```bash
   # Generate real data from game
   cd ../bpsr_labs
   poetry run bpsr-labs trade-decode capture.bin market_listings.json
   cp market_listings.json ../BPSR-API-2.0-main/
   ```

2. **View in browser:**
   - http://localhost:8989/api/market/listings
   - http://localhost:8989/api/items/search?q=Luno

3. **Set up auto-reload** to keep data fresh

---

**Still stuck? Run the diagnostic test above and share the output!**
