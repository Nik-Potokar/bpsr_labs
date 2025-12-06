# Market API Usage Guide

## Overview

The BPSR API now includes comprehensive market/trading center endpoints that provide real-time market listings, item information, and price history tracking.

## Quick Start

### 1. Start the API Server

```bash
# Basic usage (uses default market_listings.json)
./BPSR-API-2.0.exe --network auto

# Custom market data file location
./BPSR-API-2.0.exe --network auto --marketData /path/to/market_data.json

# Custom reload interval (default: 5 minutes)
./BPSR-API-2.0.exe --network auto --marketReload 10
```

### 2. Generate Market Data with bpsr-labs

```bash
# Capture packets and decode trading center data
cd /path/to/bpsr_labs
poetry run bpsr-labs trade-decode capture.bin market_listings.json

# Copy to API directory
cp market_listings.json /path/to/BPSR-API-2.0/
```

### 3. Automated Data Pipeline (Recommended)

Create a script to automatically capture and decode market data:

**Windows (PowerShell):**
```powershell
# capture_market.ps1
while ($true) {
    # Capture packets for 5 minutes
    tcpdump -i eth0 -w temp_capture.bin

    # Decode with bpsr-labs
    poetry run bpsr-labs trade-decode temp_capture.bin market_listings.json

    # Wait 5 minutes
    Start-Sleep -Seconds 300
}
```

**Linux/Mac:**
```bash
#!/bin/bash
# capture_market.sh
while true; do
    # Capture packets for 5 minutes
    timeout 300 tcpdump -i eth0 -w temp_capture.bin

    # Decode with bpsr-labs
    poetry run bpsr-labs trade-decode temp_capture.bin market_listings.json

    # Wait 5 minutes
    sleep 300
done
```

## API Endpoints

### Base URL
```
http://localhost:8989
```

### 1. Get Market Listings

**Endpoint:** `GET /api/market/listings`

**Description:** Retrieve current market listings with optional filters

**Parameters:**
- `item_id` (optional): Filter by specific item ID
- `min_price` (optional): Minimum price filter (in Luno)
- `max_price` (optional): Maximum price filter (in Luno)

**Examples:**

```bash
# Get all listings
curl http://localhost:8989/api/market/listings

# Get listings for specific item
curl "http://localhost:8989/api/market/listings?item_id=10002"

# Get listings in price range
curl "http://localhost:8989/api/market/listings?min_price=1000&max_price=5000"

# Combine filters
curl "http://localhost:8989/api/market/listings?item_id=10002&min_price=4000"
```

**Response:**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "listings": [
      {
        "listing_guid": "9d9dfe8c-55a2-4fb9-b903-1b63a4514c52",
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

### 2. Get Item Details

**Endpoint:** `GET /api/items/:id`

**Description:** Get information about a specific item

**Parameters:**
- `id` (path): Item ID

**Example:**

```bash
curl http://localhost:8989/api/items/10002
```

**Response:**
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

---

### 3. Search Items

**Endpoint:** `GET /api/items/search`

**Description:** Search for items by name (case-insensitive substring match)

**Parameters:**
- `q` (query): Search query string

**Example:**

```bash
curl "http://localhost:8989/api/items/search?q=Luno"
curl "http://localhost:8989/api/items/search?q=Rose"
```

**Response:**
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

**Note:** Limited to 50 results

---

### 4. Get Price History

**Endpoint:** `GET /api/market/prices/:id`

**Description:** Get historical price data and statistics for an item

**Parameters:**
- `id` (path): Item ID

**Example:**

```bash
curl http://localhost:8989/api/market/prices/10002
```

**Response:**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "item_id": 10002,
    "item_name": "Luno",
    "history": [
      {
        "item_id": 10002,
        "price_luno": 4500,
        "quantity": 3,
        "recorded_at": 1733493600
      },
      {
        "item_id": 10002,
        "price_luno": 4200,
        "quantity": 5,
        "recorded_at": 1733490000
      }
    ],
    "stats": {
      "average": 4350,
      "min": 4200,
      "max": 4500,
      "count": 2
    }
  }
}
```

**Note:** Price history keeps up to 1000 records per item

---

## Error Responses

### 400 Bad Request
```json
{
  "code": 1,
  "msg": "Invalid item ID"
}
```

### 404 Not Found
```json
{
  "code": 1,
  "msg": "Item not found"
}
```

## Command Line Options

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--network` | string | (required) | Network card description, "auto" for automatic |
| `--port` | int | 8989 | API server port |
| `--expire` | int | 10 | Data expiration time in seconds |
| `--autoCheckTime` | int | 3 | Auto network detection wait time |
| `--marketData` | string | market_listings.json | Path to market data file |
| `--marketReload` | int | 5 | Market data reload interval (minutes) |

## Integration with bpsr-labs

### Manual Workflow

```bash
# Step 1: Capture game packets (run while playing)
tcpdump -i <network_interface> -w capture.bin

# Step 2: Decode with bpsr-labs
cd /path/to/bpsr_labs
poetry run bpsr-labs trade-decode capture.bin market_listings.json

# Step 3: Copy to API directory
cp market_listings.json /path/to/BPSR-API-2.0/

# Step 4: API will auto-reload within 5 minutes (or use --marketReload flag)
```

### Automated Cron Job (Linux/Mac)

```bash
# Edit crontab
crontab -e

# Add entry to run every 10 minutes
*/10 * * * * cd /path/to/bpsr_labs && poetry run bpsr-labs trade-decode /tmp/capture.bin /path/to/BPSR-API-2.0/market_listings.json
```

### Automated Task Scheduler (Windows)

1. Open Task Scheduler
2. Create new task
3. Set trigger: Every 10 minutes
4. Set action: Run PowerShell script
5. Script:
```powershell
cd C:\path\to\bpsr_labs
poetry run bpsr-labs trade-decode C:\temp\capture.bin C:\path\to\BPSR-API-2.0\market_listings.json
```

## Data Flow Architecture

```
Game Network Traffic
        ↓
  Packet Capture (tcpdump)
        ↓
  Binary Capture File (.bin)
        ↓
  bpsr-labs decoder
        ↓
  market_listings.json
        ↓
  BPSR-API-2.0 (auto-loads every 5 min)
        ↓
  In-Memory Cache
        ↓
  REST API Endpoints
        ↓
  Client Application
```

## Client Examples

### JavaScript (Browser)

```javascript
// Get all market listings
fetch('http://localhost:8989/api/market/listings')
  .then(res => res.json())
  .then(data => {
    console.log(`Found ${data.data.count} listings`);
    data.data.listings.forEach(listing => {
      console.log(`${listing.item_name}: ${listing.price_luno} Luno`);
    });
  });

// Search for items
async function searchItem(query) {
  const res = await fetch(`http://localhost:8989/api/items/search?q=${query}`);
  const data = await res.json();
  return data.data.results;
}

// Get price history
async function getPriceHistory(itemId) {
  const res = await fetch(`http://localhost:8989/api/market/prices/${itemId}`);
  const data = await res.json();
  return {
    history: data.data.history,
    stats: data.data.stats
  };
}
```

### Python

```python
import requests

BASE_URL = "http://localhost:8989"

# Get all listings
response = requests.get(f"{BASE_URL}/api/market/listings")
listings = response.json()['data']['listings']

# Filter by item
response = requests.get(f"{BASE_URL}/api/market/listings", params={"item_id": 10002})
luno_listings = response.json()['data']['listings']

# Search items
response = requests.get(f"{BASE_URL}/api/items/search", params={"q": "Luno"})
results = response.json()['data']['results']

# Get price history
response = requests.get(f"{BASE_URL}/api/market/prices/10002")
history = response.json()['data']
print(f"Average price: {history['stats']['average']} Luno")
```

### Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type ListingsResponse struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data struct {
        Listings []struct {
            ItemID    int    `json:"item_id"`
            ItemName  string `json:"item_name"`
            PriceLuno int    `json:"price_luno"`
            Quantity  int    `json:"quantity"`
        } `json:"listings"`
        Count int `json:"count"`
    } `json:"data"`
}

func getMarketListings() (*ListingsResponse, error) {
    resp, err := http.Get("http://localhost:8989/api/market/listings")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ListingsResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    listings, err := getMarketListings()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d listings\n", listings.Data.Count)
    for _, listing := range listings.Data.Listings {
        fmt.Printf("%s: %d Luno (qty: %d)\n",
            listing.ItemName, listing.PriceLuno, listing.Quantity)
    }
}
```

## Troubleshooting

### No market data showing

**Problem:** API returns empty listings
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

**Solutions:**
1. Check if `market_listings.json` exists in the API directory
2. Check API logs for file loading errors
3. Verify JSON file format matches expected structure
4. Wait for next reload cycle (default: 5 minutes)

### Item names missing

**Problem:** Items show ID but no name

**Solutions:**
1. Verify `items.json` was embedded correctly
2. Check API startup logs for "Loaded X items into catalog"
3. Rebuild the API executable if items.json was updated

### Price history not available

**Problem:** `/api/market/prices/:id` returns 404

**Cause:** Price history builds over time as listings are imported

**Solution:** Wait for multiple data reloads to accumulate history

## Performance Notes

- **Item Catalog**: Loaded once at startup (6000+ items)
- **Market Listings**: Reloaded every 5 minutes (configurable)
- **Price History**: Kept in memory (max 1000 records per item)
- **Old Listings**: Auto-cleaned after 1 hour
- **Concurrent Requests**: Thread-safe with RWMutex locks

## Next Steps

1. **Build Data Pipeline**: Set up automated packet capture and decoding
2. **Create Dashboard**: Build web UI to visualize market data
3. **Add Alerts**: Implement price monitoring and notifications
4. **Export Data**: Add CSV/Excel export endpoints
5. **Analytics**: Calculate trends, predictions, and market insights

---

**Created:** 2025-12-06
**Version:** 1.0
**Author:** Claude Code
