# Market/Trading Center Integration Plan for BPSR-API-2.0

**Date**: 2025-12-06
**Status**: Ready for Implementation
**Repositories**:
- Data Source: `bpsr_labs` (Python - Packet decoder & item catalog)
- API Target: `BPSR-API-2.0` (Go/Golang - REST API)

---

## Executive Summary

**Goal**: Add market/trading center functionality to BPSR-API-2.0 using data from bpsr_labs

**Current State**:
- ✅ bpsr_labs has trading center decoder and 6K+ item catalog
- ✅ BPSR-API-2.0 has packet capture infrastructure
- ❌ BPSR-API-2.0 has NO market endpoints yet

**Strategy**: Add market packet processing to BPSR-API-2.0 using similar approach to existing enemy/scene data

---

## Architecture Overview

### Current BPSR-API-2.0 Architecture

```
Network Traffic
    ↓
npcap (Packet Capture)
    ↓
cap_device.go (TCP Reassembly)
    ↓
Protobuf Decoder (bp.pb.go)
    ↓
In-Memory Cache (global/cache.go)
    ↓
Gin REST API (main.go)
    ↓
JSON Response
```

### Proposed Market Integration

```
Same Packet Capture Pipeline
    ↓
NEW: Market Packet Detector (identify ExchangeNoticeDetail packets)
    ↓
NEW: Market Data Parser (decode listing data)
    ↓
NEW: Market Cache (similar to SceneMonsterList)
    ↓
NEW: Market API Endpoints
    ↓
JSON Response with Market Data
```

---

## Implementation Plan

### Phase 1: Data Preparation (30 minutes)

#### Task 1.1: Import Item Catalog

**Location**: `BPSR-API-2.0-main/global/`

Create `items.go`:
```go
package global

import (
    _ "embed"
    "encoding/json"
    "log"
    "sync"
)

//go:embed items.json
var itemsJSON []byte

var ItemCatalog = make(map[int]string)
var ItemCatalogLock = sync.RWMutex{}

func InitItemCatalog() {
    ItemCatalogLock.Lock()
    defer ItemCatalogLock.Unlock()

    err := json.Unmarshal(itemsJSON, &ItemCatalog)
    if err != nil {
        log.Println("Warning: Failed to load item catalog:", err)
        return
    }
    log.Printf("Loaded %d items into catalog", len(ItemCatalog))
}

func GetItemName(itemID int) string {
    ItemCatalogLock.RLock()
    defer ItemCatalogLock.RUnlock()

    if name, ok := ItemCatalog[itemID]; ok {
        return name
    }
    return ""
}
```

**Action Required**:
1. Copy `bpsr_labs/data/game-data/item_name_map.json` to `BPSR-API-2.0-main/global/items.json`
2. Add `InitItemCatalog()` call in `main.go` (after line 72, alongside `InitMonsterNames()`)

---

### Phase 2: Market Data Structures (1 hour)

#### Task 2.1: Add Market Types

**Location**: `BPSR-API-2.0-main/global/cache.go`

Add to the file:
```go
// Market listing data
var MarketListings = make(map[string]*MarketListing)
var MarketListingsLock = sync.RWMutex{}

type MarketListing struct {
    ListingGUID  string  `json:"listing_guid"`
    ItemID       int     `json:"item_id"`
    ItemName     string  `json:"item_name"`
    PriceLuno    int     `json:"price_luno"`
    Quantity     int     `json:"quantity"`
    BindFlag     bool    `json:"bind_flag"`
    ListedAt     int64   `json:"listed_at"`      // Unix timestamp
    CapturedAt   int64   `json:"captured_at"`    // When we captured this
    UnitPrice    int     `json:"unit_price"`     // Price per item
}

// Price history for analytics
var PriceHistory = make(map[int][]PriceRecord)
var PriceHistoryLock = sync.RWMutex{}

type PriceRecord struct {
    ItemID     int   `json:"item_id"`
    PriceLuno  int   `json:"price_luno"`
    Quantity   int   `json:"quantity"`
    RecordedAt int64 `json:"recorded_at"`
}

func AddMarketListing(listing *MarketListing) {
    MarketListingsLock.Lock()
    defer MarketListingsLock.Unlock()

    // Add to current listings
    MarketListings[listing.ListingGUID] = listing

    // Record in price history
    PriceHistoryLock.Lock()
    defer PriceHistoryLock.Unlock()

    history := PriceHistory[listing.ItemID]
    history = append(history, PriceRecord{
        ItemID:     listing.ItemID,
        PriceLuno:  listing.PriceLuno,
        Quantity:   listing.Quantity,
        RecordedAt: listing.CapturedAt,
    })

    // Keep only last 1000 records per item
    if len(history) > 1000 {
        history = history[len(history)-1000:]
    }

    PriceHistory[listing.ItemID] = history
}

func CleanOldListings(maxAge int64) {
    MarketListingsLock.Lock()
    defer MarketListingsLock.Unlock()

    now := time.Now().Unix()
    for guid, listing := range MarketListings {
        if now-listing.CapturedAt > maxAge {
            delete(MarketListings, guid)
        }
    }
}
```

---

### Phase 3: Packet Processing (2-3 hours)

#### Task 3.1: Identify Market Packets

**Challenge**: Need to identify the packet signature for trading center data

**Options**:

**Option A: Use bpsr_labs protobuf definitions**
- Copy protobuf definitions from `bpsr_labs` StarResonanceData
- Generate Go code: `protoc --go_out=. exchange.proto`
- Integrate into existing `pb/bp.pb.go`

**Option B: Pattern matching approach** (faster, simpler)
- Use same approach as bpsr_labs V1 decoder
- Look for specific byte patterns in packets
- Extract listing data heuristically

**Recommended**: Start with Option B (faster to implement), migrate to Option A later

#### Task 3.2: Add Market Packet Handler

**Location**: `BPSR-API-2.0-main/ncap/cap_device.go`

Add new method:
```go
func (cd *CapDevice) handleMarketPacket(data []byte) {
    // Pattern matching for market listings
    // Based on bpsr_labs trading_center_decode.py logic

    // Look for ExchangeNoticeDetail signature
    // Packet type 0x0006 (FrameDown)
    if len(data) < 6 {
        return
    }

    // Extract packet type
    pktType := binary.BigEndian.Uint16(data[4:6])
    if pktType != 0x0006 && pktType != 0x8006 { // 0x8006 = compressed
        return
    }

    // Check if compressed (zstd)
    isCompressed := (pktType & 0x8000) != 0
    body := data[6:]

    if isCompressed && len(body) > 4 {
        // Decompress with zstd
        decoder, _ := zstd.NewReader(nil)
        decompressed, err := decoder.DecodeAll(body[4:], nil)
        if err == nil {
            body = decompressed
        }
    }

    // Parse listings from body
    cd.parseMarketListings(body)
}

func (cd *CapDevice) parseMarketListings(data []byte) {
    // Iterate through protobuf messages in data
    offset := 0
    for offset < len(data) {
        // Look for field tag 0x0A (varint for message field)
        if data[offset] != 0x0A {
            offset++
            continue
        }

        // Read varint length
        msgLen, n := protowire.ConsumeVarint(data[offset+1:])
        if n <= 0 {
            offset++
            continue
        }

        payloadStart := offset + 1 + n
        payloadEnd := payloadStart + int(msgLen)

        if payloadEnd > len(data) {
            break
        }

        payload := data[payloadStart:payloadEnd]

        // Try to parse as market listing
        cd.decodeMarketListing(payload)

        offset = payloadEnd
    }
}

func (cd *CapDevice) decodeMarketListing(data []byte) {
    // Parse protobuf fields
    // This is simplified - real implementation needs full protobuf parsing

    var itemID int
    var price int
    var quantity int
    var guid string
    var timestamp int64

    // Parse protobuf fields (simplified example)
    // Real implementation should use generated protobuf code
    offset := 0
    for offset < len(data) {
        if offset+1 >= len(data) {
            break
        }

        tag := data[offset]
        offset++

        switch tag {
        case 0x08: // price field
            val, n := protowire.ConsumeVarint(data[offset:])
            if n > 0 {
                price = int(val)
                offset += n
            }
        case 0x10: // quantity field
            val, n := protowire.ConsumeVarint(data[offset:])
            if n > 0 {
                quantity = int(val)
                offset += n
            }
        // Add more fields as needed
        }
    }

    // Create listing object
    if itemID > 0 && price > 0 {
        listing := &global.MarketListing{
            ListingGUID: guid,
            ItemID:      itemID,
            ItemName:    global.GetItemName(itemID),
            PriceLuno:   price,
            Quantity:    quantity,
            ListedAt:    timestamp,
            CapturedAt:  time.Now().Unix(),
            UnitPrice:   price / quantity,
        }

        global.AddMarketListing(listing)
    }
}
```

**Integration point**: Add call to `handleMarketPacket()` in existing packet processing loop

---

### Phase 4: API Endpoints (1 hour)

#### Task 4.1: Add Market Endpoints

**Location**: `BPSR-API-2.0-main/main.go`

Add these routes in `Openapi()` function (after line 182):

```go
// Get all market listings
s.GET("/api/market/listings", func(ctx *gin.Context) {
    global.MarketListingsLock.RLock()
    defer global.MarketListingsLock.RUnlock()

    // Optional filters
    itemIDStr := ctx.Query("item_id")
    minPriceStr := ctx.Query("min_price")
    maxPriceStr := ctx.Query("max_price")

    listings := make([]*global.MarketListing, 0)
    for _, listing := range global.MarketListings {
        // Apply filters
        if itemIDStr != "" {
            itemID, _ := strconv.Atoi(itemIDStr)
            if listing.ItemID != itemID {
                continue
            }
        }

        if minPriceStr != "" {
            minPrice, _ := strconv.Atoi(minPriceStr)
            if listing.PriceLuno < minPrice {
                continue
            }
        }

        if maxPriceStr != "" {
            maxPrice, _ := strconv.Atoi(maxPriceStr)
            if listing.PriceLuno > maxPrice {
                continue
            }
        }

        listings = append(listings, listing)
    }

    ctx.JSON(200, gin.H{
        "code": 0,
        "msg":  "OK",
        "data": gin.H{
            "listings": listings,
            "count":    len(listings),
        },
    })
})

// Get item details
s.GET("/api/items/:id", func(ctx *gin.Context) {
    itemIDStr := ctx.Param("id")
    itemID, err := strconv.Atoi(itemIDStr)
    if err != nil {
        ctx.JSON(400, gin.H{
            "code": 1,
            "msg":  "Invalid item ID",
        })
        return
    }

    itemName := global.GetItemName(itemID)
    if itemName == "" {
        ctx.JSON(404, gin.H{
            "code": 1,
            "msg":  "Item not found",
        })
        return
    }

    ctx.JSON(200, gin.H{
        "code": 0,
        "msg":  "OK",
        "data": gin.H{
            "item_id":   itemID,
            "item_name": itemName,
        },
    })
})

// Get price history for an item
s.GET("/api/market/prices/:id", func(ctx *gin.Context) {
    itemIDStr := ctx.Param("id")
    itemID, err := strconv.Atoi(itemIDStr)
    if err != nil {
        ctx.JSON(400, gin.H{
            "code": 1,
            "msg":  "Invalid item ID",
        })
        return
    }

    global.PriceHistoryLock.RLock()
    defer global.PriceHistoryLock.RUnlock()

    history, exists := global.PriceHistory[itemID]
    if !exists || len(history) == 0 {
        ctx.JSON(404, gin.H{
            "code": 1,
            "msg":  "No price history available",
        })
        return
    }

    // Calculate statistics
    var total int
    var min = history[0].PriceLuno
    var max = history[0].PriceLuno

    for _, record := range history {
        total += record.PriceLuno
        if record.PriceLuno < min {
            min = record.PriceLuno
        }
        if record.PriceLuno > max {
            max = record.PriceLuno
        }
    }

    avg := total / len(history)

    ctx.JSON(200, gin.H{
        "code": 0,
        "msg":  "OK",
        "data": gin.H{
            "item_id":   itemID,
            "item_name": global.GetItemName(itemID),
            "history":   history,
            "stats": gin.H{
                "average": avg,
                "min":     min,
                "max":     max,
                "count":   len(history),
            },
        },
    })
})

// Search items by name
s.GET("/api/items/search", func(ctx *gin.Context) {
    query := ctx.Query("q")
    if query == "" {
        ctx.JSON(400, gin.H{
            "code": 1,
            "msg":  "Missing search query",
        })
        return
    }

    global.ItemCatalogLock.RLock()
    defer global.ItemCatalogLock.RUnlock()

    results := make([]gin.H, 0)
    query = strings.ToLower(query)

    for itemID, itemName := range global.ItemCatalog {
        if strings.Contains(strings.ToLower(itemName), query) {
            results = append(results, gin.H{
                "item_id":   itemID,
                "item_name": itemName,
            })

            // Limit to 50 results
            if len(results) >= 50 {
                break
            }
        }
    }

    ctx.JSON(200, gin.H{
        "code": 0,
        "msg":  "OK",
        "data": gin.H{
            "results": results,
            "count":   len(results),
        },
    })
})
```

---

## Alternative: Hybrid Approach (Recommended)

Instead of implementing packet parsing in Go, use bpsr_labs as a **data pipeline**:

### Hybrid Architecture

```
Game Network → bpsr_labs decoder → JSON files → BPSR-API reads files
```

**Advantages**:
- ✅ No need to reimplement packet parsing in Go
- ✅ Use proven bpsr_labs decoder
- ✅ Faster to implement
- ✅ Easier to maintain

**Implementation**:

1. **Run bpsr_labs in background** (scheduled task every 5-15 minutes):
   ```bash
   # Capture packets
   tcpdump -i eth0 -w capture.bin

   # Decode with bpsr_labs
   bpsr-labs trade-decode capture.bin /tmp/market_listings.json
   ```

2. **BPSR-API reads JSON files**:
   ```go
   // In global/market_loader.go
   func LoadMarketDataFromFile(path string) error {
       data, err := ioutil.ReadFile(path)
       if err != nil {
           return err
       }

       var listings []struct {
           PriceLuno int    `json:"price_luno"`
           Quantity  int    `json:"quantity"`
           ItemID    int    `json:"item_id"`
           ItemName  string `json:"item_name"`
           Metadata  struct {
               RawEntry struct {
                   GUID       string `json:"guid"`
                   NoticeTime int64  `json:"noticeTime"`
               } `json:"raw_entry"`
           } `json:"metadata"`
       }

       err = json.Unmarshal(data, &listings)
       if err != nil {
           return err
       }

       // Import to cache
       for _, item := range listings {
           listing := &MarketListing{
               ListingGUID: item.Metadata.RawEntry.GUID,
               ItemID:      item.ItemID,
               ItemName:    item.ItemName,
               PriceLuno:   item.PriceLuno,
               Quantity:    item.Quantity,
               ListedAt:    item.Metadata.RawEntry.NoticeTime,
               CapturedAt:  time.Now().Unix(),
               UnitPrice:   item.PriceLuno / item.Quantity,
           }
           AddMarketListing(listing)
       }

       return nil
   }

   // Periodic reload (every 5 minutes)
   func StartMarketDataReloader() {
       ticker := time.NewTicker(5 * time.Minute)
       go func() {
           for range ticker.C {
               err := LoadMarketDataFromFile("/tmp/market_listings.json")
               if err != nil {
                   log.Println("Failed to reload market data:", err)
               } else {
                   log.Println("Market data reloaded successfully")
               }
           }
       }()
   }
   ```

3. **Call in main.go**:
   ```go
   // After line 72
   global.StartMarketDataReloader()
   ```

---

## Testing Plan

### Phase 1: Item Catalog
```bash
# Start API
./BPSR-API-2.0.exe

# Test item lookup
curl http://localhost:8989/api/items/10002
# Expected: {"code":0,"msg":"OK","data":{"item_id":10002,"item_name":"Luno"}}

# Test item search
curl "http://localhost:8989/api/items/search?q=sword"
# Expected: List of items containing "sword"
```

### Phase 2: Market Listings
```bash
# Get all listings
curl http://localhost:8989/api/market/listings
# Expected: List of current market listings

# Filter by item
curl "http://localhost:8989/api/market/listings?item_id=12345"
# Expected: Listings for item 12345 only

# Price history
curl http://localhost:8989/api/market/prices/12345
# Expected: Price history with min/max/avg stats
```

---

## Timeline

| Phase | Task | Time | Dependency |
|-------|------|------|------------|
| 1 | Copy item catalog | 15 min | None |
| 1 | Add item catalog loader | 15 min | Phase 1.1 |
| 2 | Add market data structures | 30 min | Phase 1 |
| 3A | Implement hybrid file loader | 1 hour | Phase 2 |
| 3B | Set up bpsr_labs pipeline | 30 min | None |
| 4 | Add API endpoints | 1 hour | Phase 3 |
| 5 | Testing | 1 hour | Phase 4 |
| **Total** | **~5 hours** | | |

**Recommended Approach**: Hybrid (faster, more reliable)

---

## File Changes Summary

### New Files
- `BPSR-API-2.0-main/global/items.go` - Item catalog
- `BPSR-API-2.0-main/global/items.json` - Item data (copied from bpsr_labs)
- `BPSR-API-2.0-main/global/market_loader.go` - Market data file loader (hybrid approach)

### Modified Files
- `BPSR-API-2.0-main/global/cache.go` - Add market structures
- `BPSR-API-2.0-main/main.go` - Add market endpoints
- `BPSR-API-2.0-main/go.mod` - May need additional dependencies

### Optional (Native approach)
- `BPSR-API-2.0-main/ncap/cap_device.go` - Add market packet handler

---

## Next Steps

1. **Choose approach**:
   - Hybrid (recommended): Faster, uses proven decoder
   - Native: Full integration, real-time updates

2. **Start with Phase 1**: Import item catalog (15 minutes)

3. **Test item endpoints**: Verify catalog is working

4. **Implement market data loading**: Choose hybrid or native

5. **Add API endpoints**: Expose market data

6. **Test and iterate**: Verify all functionality works

---

## Questions?

- Which approach do you prefer? (Hybrid vs Native)
- Do you want to start with Phase 1?
- Any specific features or requirements?

---

**Author**: Claude Code
**Date**: 2025-12-06
