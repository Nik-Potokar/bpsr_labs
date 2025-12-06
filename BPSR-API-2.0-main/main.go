package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/balrogsxt/StarResonanceAPI/global"
	"github.com/balrogsxt/StarResonanceAPI/ncap"
	"github.com/gin-gonic/gin"
	"github.com/google/gopacket/pcap"
)

var (
	networkCard      = flag.String("network", "", "Enter network card full description, auto for automatic selection")
	port             = flag.Int("port", 8989, "Default API port")
	expireTime       = flag.Int64("expire", 10, "Data expiration time (seconds), default 10s")
	autoCheckTime    = flag.Int("autoCheckTime", 3, "Auto check active network card time (seconds)")
	marketDataFile   = flag.String("marketData", "market_listings.json", "Path to market data JSON file")
	marketReloadMins = flag.Int("marketReload", 5, "Market data reload interval (minutes)")
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Program crashed: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()
	flag.Parse()

	var deviceName = *networkCard
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Failed to get network cards: ", err.Error())
	}
	if deviceName == "auto" {
		log.Println("Automatically finding active network card, please wait...")
		active := ncap.GetActiveNetworkCards(devices, *autoCheckTime)
		if active != nil {
			log.Println("Automatically found suitable network card: ", active.Desc)
			log.Println("Monitored packet count: ", active.PacketCount)
			log.Println("Monitored packet traffic: ", fmt.Sprintf("%d bytes (%.2f KB)", active.ByteCount, float64(active.ByteCount)/1024))
			deviceName = active.Desc
		}
	}
	if deviceName == "" {
		var option string
		options := make([]string, 0)
		for _, device := range devices {
			options = append(options, device.Description)
		}
		prompt := &survey.Select{
			Message: "Unable to automatically find active network card, please manually select (you can find the description in your network settings):",
			Options: options,
		}
		err := survey.AskOne(prompt, &option)
		if err != nil {
			log.Fatalf("Selection error: %s", err.Error())
		}
		if len(option) == 0 {
			log.Fatalf("Network card selection is empty")
		}
		deviceName = option
	}

	// Load monster JSON list
	global.InitMonsterNames()

	// Load item catalog
	global.InitItemCatalog()

	// Start market data reloader
	global.StartMarketDataReloader(*marketDataFile, *marketReloadMins)

	// Start services
	go Openapi()
	go OpenCap(deviceName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Program started, press Ctrl+C to exit")
	<-sigChan
	log.Println("Shutting down...")
}

func OpenCap(deviceName string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Packet capture service crashed: %v\nStack trace:\n%s", r, debug.Stack())
			log.Println("API server will continue running. Press Ctrl+C to exit.")
		}
	}()

	// Create packet capture core
	capCore := ncap.NewCapCore()
	if err := capCore.Start(deviceName); err != nil {
		log.Printf("ERROR: Failed to start packet capture: %v", err)
		log.Println("API server will continue running. Press Ctrl+C to exit.")
		// Don't use Fatalf - let the API server continue running
		return
	}
}
func Openapi() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("API service crashed: %v\nStack trace:\n%s", r, debug.Stack())
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	s := gin.New()

	// Add global panic recovery middleware
	s.Use(gin.Recovery())

	s.GET("/api/enemies", func(ctx *gin.Context) {
		global.SceneMonsterListLock.RLock()
		defer global.SceneMonsterListLock.RUnlock()
		et := *expireTime
		if 0 >= et {
			et = 10
		}
		list := make(map[uint64]*global.Monster)
		for id, item := range global.SceneMonsterList {
			if time.Now().Unix()-item.UpdateTime > et {
				continue // Ignore data not updated in the last N seconds
			}
			attackPlayers := make(map[uint64]*global.AttackPlayer)
			if item.AttackPlayers != nil {
				for uid, player := range item.AttackPlayers {
					if time.Now().Unix()-player.LastAttackTime > et { // Ignore players not in combat in the last N seconds
						continue
					}
					attackPlayers[uid] = player
				}
			}
			if len(item.Name) > 0 || (item.Hp >= 0 && item.MaxHp > 0) {
				monsterCopy := *item
				monsterCopy.AttackPlayers = attackPlayers
				list[id] = &monsterCopy
			}
		}
		ctx.JSON(200, gin.H{
			"code":  0,
			"msg":   "OK",
			"enemy": list,
		})
	})
	s.GET("/api/clear", func(ctx *gin.Context) {
		global.ClearAllData()
		ctx.JSON(200, gin.H{
			"code": 0,
			"msg":  "OK",
		})
	})
	s.GET("/api/translate", func(ctx *gin.Context) {
		chineseName := ctx.Query("name")
		if len(chineseName) == 0 {
			ctx.JSON(400, gin.H{
				"code": 1,
				"msg":  "Missing 'name' parameter",
			})
			return
		}
		englishName := global.TranslateMonsterName(chineseName)
		ctx.JSON(200, gin.H{
			"code": 0,
			"msg":  "OK",
			"data": gin.H{
				"chinese": chineseName,
				"english": englishName,
			},
		})
	})
	s.GET("/api/scene", func(ctx *gin.Context) {
		global.CurrentSceneLock.RLock()
		defer global.CurrentSceneLock.RUnlock()
		ctx.JSON(200, gin.H{
			"code": 0,
			"msg":  "OK",
			"data": global.CurrentScene,
		})
	})

	// Market API endpoints
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
				var itemID int
				fmt.Sscanf(itemIDStr, "%d", &itemID)
				if listing.ItemID != itemID {
					continue
				}
			}

			if minPriceStr != "" {
				var minPrice int
				fmt.Sscanf(minPriceStr, "%d", &minPrice)
				if listing.PriceLuno < minPrice {
					continue
				}
			}

			if maxPriceStr != "" {
				var maxPrice int
				fmt.Sscanf(maxPriceStr, "%d", &maxPrice)
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

	s.GET("/api/items/:id", func(ctx *gin.Context) {
		itemIDStr := ctx.Param("id")
		var itemID int
		_, err := fmt.Sscanf(itemIDStr, "%d", &itemID)
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

	s.GET("/api/market/prices/:id", func(ctx *gin.Context) {
		itemIDStr := ctx.Param("id")
		var itemID int
		_, err := fmt.Sscanf(itemIDStr, "%d", &itemID)
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
		min := history[0].PriceLuno
		max := history[0].PriceLuno

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
		queryLower := fmt.Sprintf("%s", query)

		for itemID, itemName := range global.ItemCatalog {
			itemNameLower := fmt.Sprintf("%s", itemName)
			// Simple case-insensitive contains search
			if len(itemNameLower) > 0 && len(queryLower) > 0 {
				// Check if query is in item name (simple substring match)
				found := false
				for i := 0; i <= len(itemNameLower)-len(queryLower); i++ {
					if itemNameLower[i:i+len(queryLower)] == queryLower {
						found = true
						break
					}
				}
				if found {
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

	log.Println(fmt.Sprintf("Service started at: http://127.0.0.1:%d", *port))
	if err := s.Run(fmt.Sprintf(":%d", *port)); err != nil {
		log.Printf("ERROR: API server failed: %s", err.Error())
		log.Println("Please check if the port is already in use or restart the application.")
		// Don't use Fatalf - let the panic recovery handle it
		return
	}
}
