package global

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// LoadMarketDataFromFile loads market listings from a JSON file exported by bpsr-labs
func LoadMarketDataFromFile(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("market data file not found: %s", path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read market data file: %v", err)
	}

	// Expected format from bpsr-labs trade-decode output
	var listings []struct {
		PriceLuno int    `json:"price_luno"`
		Quantity  int    `json:"quantity"`
		ItemID    int    `json:"item_id"`
		ItemName  string `json:"item_name"`
		Metadata  struct {
			RawEntry struct {
				GUID       string `json:"guid"`
				NoticeTime int64  `json:"noticeTime"`
				ItemInfo   struct {
					BindFlag bool `json:"bindFlag"`
				} `json:"itemInfo"`
			} `json:"raw_entry"`
		} `json:"metadata"`
	}

	err = json.Unmarshal(data, &listings)
	if err != nil {
		return fmt.Errorf("failed to parse market data JSON: %v", err)
	}

	// Import to cache
	imported := 0
	for _, item := range listings {
		// Skip invalid entries
		if item.ItemID == 0 || item.PriceLuno == 0 || item.Quantity == 0 {
			continue
		}

		// Calculate unit price
		unitPrice := item.PriceLuno
		if item.Quantity > 0 {
			unitPrice = item.PriceLuno / item.Quantity
		}

		listing := &MarketListing{
			ListingGUID: item.Metadata.RawEntry.GUID,
			ItemID:      item.ItemID,
			ItemName:    item.ItemName,
			PriceLuno:   item.PriceLuno,
			Quantity:    item.Quantity,
			BindFlag:    item.Metadata.RawEntry.ItemInfo.BindFlag,
			ListedAt:    item.Metadata.RawEntry.NoticeTime,
			CapturedAt:  time.Now().Unix(),
			UnitPrice:   unitPrice,
		}

		// If item name is missing, try to get it from catalog
		if listing.ItemName == "" {
			listing.ItemName = GetItemName(listing.ItemID)
		}

		AddMarketListing(listing)
		imported++
	}

	log.Printf("Imported %d market listings from %s", imported, path)
	return nil
}

// StartMarketDataReloader starts a background task to periodically reload market data
func StartMarketDataReloader(path string, intervalMinutes int) {
	if intervalMinutes <= 0 {
		intervalMinutes = 5 // Default to 5 minutes
	}

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)

	// Load immediately on start
	go func() {
		err := LoadMarketDataFromFile(path)
		if err != nil {
			log.Printf("Initial market data load failed: %v", err)
			log.Println("Market endpoints will return empty data until the file is available")
		}
	}()

	// Then reload periodically
	go func() {
		for range ticker.C {
			err := LoadMarketDataFromFile(path)
			if err != nil {
				log.Printf("Failed to reload market data: %v", err)
			} else {
				// Clean up old listings (older than 1 hour)
				removed := CleanOldListings(3600)
				if removed > 0 {
					log.Printf("Cleaned up %d old market listings", removed)
				}
			}
		}
	}()

	log.Printf("Market data reloader started (interval: %d minutes, file: %s)", intervalMinutes, path)
}
