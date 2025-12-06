# Testing the Market API - Quick Guide

## ✅ Pre-Build Test (Check Code Syntax)

Since we can't build in this environment due to network restrictions, here's how to verify everything works when you build it on your machine:

## 🔨 Building the API (On Your Machine)

```bash
cd BPSR-API-2.0-main

# Build the executable
go build -o BPSR-API-2.0.exe .

# Or for Linux/Mac
go build -o BPSR-API-2.0 .
```

**Expected Output:**
- No errors
- Binary file created: `BPSR-API-2.0.exe` (Windows) or `BPSR-API-2.0` (Linux/Mac)

---

## 🚀 Starting the API

### Option 1: Basic Start (Auto-detect network)
```bash
./BPSR-API-2.0.exe --network auto
```

### Option 2: Manual Network Selection
```bash
# The program will show you a list to choose from
./BPSR-API-2.0.exe
```

### Option 3: With Custom Settings
```bash
./BPSR-API-2.0.exe --network auto --port 8989 --marketData market_listings.json --marketReload 5
```

**Expected Output:**
```
Loaded X monster names
Loaded 6367 items into catalog
Market data reloader started (interval: 5 minutes, file: market_listings.json)
Initial market data load failed: market data file not found: market_listings.json
Market endpoints will return empty data until the file is available
Service started at: http://127.0.0.1:8989
```

---

## 🧪 Testing Endpoints (Without Market Data)

Even without market data, you can test if the API is running:

### 1. Test Item Catalog Endpoint
```bash
# Get item by ID (Luno)
curl http://localhost:8989/api/items/10002
```

**Expected Response:**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "item_id": 10002,
    "item_name": "Luno"
  }
}
```

### 2. Test Item Search
```bash
# Search for items containing "Luno"
curl "http://localhost:8989/api/items/search?q=Luno"
```

**Expected Response:**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "results": [
      {
        "item_id": 10002,
        "item_name": "Luno"
      },
      {
        "item_id": 10008,
        "item_name": "Luno (Bound)"
      }
    ],
    "count": 2
  }
}
```

### 3. Test Market Listings (Will be empty until you add data)
```bash
curl http://localhost:8989/api/market/listings
```

**Expected Response (empty data):**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "listings": [],
    "count": 0
  }
}
```

### 4. Test Existing Endpoints (Should still work)
```bash
# Scene data
curl http://localhost:8989/api/scene

# Enemies data
curl http://localhost:8989/api/enemies
```

---

## 📊 Testing WITH Market Data

### Step 1: Create Test Market Data

Create `market_listings.json` in the API directory:

```json
[
  {
    "price_luno": 4500,
    "quantity": 3,
    "item_id": 10002,
    "item_name": "Luno",
    "metadata": {
      "raw_entry": {
        "guid": "test-guid-001",
        "noticeTime": 1733493600,
        "itemInfo": {
          "bindFlag": false
        }
      }
    }
  },
  {
    "price_luno": 12000,
    "quantity": 1,
    "item_id": 10003,
    "item_name": "Rose Orb",
    "metadata": {
      "raw_entry": {
        "guid": "test-guid-002",
        "noticeTime": 1733494200,
        "itemInfo": {
          "bindFlag": false
        }
      }
    }
  }
]
```

### Step 2: Wait for Auto-Reload or Restart

**Option A:** Wait 5 minutes for auto-reload

**Option B:** Restart the API to load immediately
```bash
# Ctrl+C to stop
# Then restart
./BPSR-API-2.0.exe --network auto
```

### Step 3: Test Market Endpoints

```bash
# Get all listings
curl http://localhost:8989/api/market/listings

# Filter by item
curl "http://localhost:8989/api/market/listings?item_id=10002"

# Price range filter
curl "http://localhost:8989/api/market/listings?min_price=4000&max_price=5000"
```

**Expected Response:**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "listings": [
      {
        "listing_guid": "test-guid-001",
        "item_id": 10002,
        "item_name": "Luno",
        "price_luno": 4500,
        "quantity": 3,
        "bind_flag": false,
        "listed_at": 1733493600,
        "captured_at": 1733497200,
        "unit_price": 1500
      }
    ],
    "count": 1
  }
}
```

---

## 🌐 Test in Browser

Simply open your browser and navigate to:

- `http://localhost:8989/api/items/10002` - Get item details
- `http://localhost:8989/api/items/search?q=Luno` - Search items
- `http://localhost:8989/api/market/listings` - View all listings
- `http://localhost:8989/api/scene` - Scene data
- `http://localhost:8989/api/enemies` - Enemy data

---

## 🔍 Troubleshooting

### Issue: "Service failed: bind: address already in use"

**Solution:** Another program is using port 8989
```bash
# Use a different port
./BPSR-API-2.0.exe --network auto --port 9000
```

### Issue: "Item not found" or empty catalog

**Problem:** Item catalog didn't load

**Check logs for:**
```
Loaded 6367 items into catalog
```

**If missing:**
- Verify `global/items.json` exists
- Rebuild the executable (it should embed the file)

### Issue: Market listings always empty

**Causes:**
1. No `market_listings.json` file exists
2. File is in wrong location
3. JSON format is incorrect
4. Haven't waited for reload cycle

**Solutions:**
```bash
# Check if file exists
ls -la market_listings.json

# Check file location matches --marketData flag
./BPSR-API-2.0.exe --network auto --marketData /full/path/to/market_listings.json

# Force immediate reload by restarting
# Ctrl+C then restart
```

### Issue: No market data file errors

**Expected on first run:**
```
Initial market data load failed: market data file not found: market_listings.json
Market endpoints will return empty data until the file is available
```

**This is normal!** Just create the file and wait for reload, or restart.

---

## ✅ Success Indicators

You'll know it's working when you see:

1. **On Startup:**
```
Loaded X monster names
Loaded 6367 items into catalog
Market data reloader started (interval: 5 minutes, file: market_listings.json)
Service started at: http://127.0.0.1:8989
```

2. **Item Endpoint Works:**
```bash
$ curl http://localhost:8989/api/items/10002
{"code":0,"msg":"OK","data":{"item_id":10002,"item_name":"Luno"}}
```

3. **Search Works:**
```bash
$ curl "http://localhost:8989/api/items/search?q=Luno"
{"code":0,"msg":"OK","data":{"count":2,"results":[...]}}
```

4. **Market Data Loads:**
```
Imported 3 market listings from market_listings.json
```

---

## 📋 Quick Test Checklist

- [ ] API builds without errors
- [ ] API starts successfully
- [ ] Item catalog loads (6367 items)
- [ ] `/api/items/:id` endpoint works
- [ ] `/api/items/search` endpoint works
- [ ] `/api/market/listings` endpoint responds (even if empty)
- [ ] Existing endpoints still work (`/api/scene`, `/api/enemies`)
- [ ] Market data file loads when created
- [ ] Auto-reload works after 5 minutes

---

## 🎯 Next Steps After Testing

Once everything works:

1. **Generate Real Market Data:**
```bash
cd /path/to/bpsr_labs
poetry run bpsr-labs trade-decode capture.bin market_listings.json
cp market_listings.json /path/to/BPSR-API-2.0-main/
```

2. **Set Up Automated Pipeline:**
- See `MARKET_API_USAGE.md` for automation scripts

3. **Build a Dashboard:**
- Use the API to create a web UI
- Visualize market trends
- Set up price alerts

---

**Need help?** Check `MARKET_API_USAGE.md` for complete documentation!
