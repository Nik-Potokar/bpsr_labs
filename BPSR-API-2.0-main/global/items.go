package global

import (
	_ "embed"
	"encoding/json"
	"log"
	"strconv"
	"sync"
)

//go:embed items.json
var itemsJSON []byte

var ItemCatalog = make(map[int]string)
var ItemCatalogLock = sync.RWMutex{}

// InitItemCatalog loads the item catalog from embedded JSON file
func InitItemCatalog() {
	ItemCatalogLock.Lock()
	defer ItemCatalogLock.Unlock()

	// Parse JSON into map[string]string first (JSON keys are strings)
	var rawCatalog map[string]string
	err := json.Unmarshal(itemsJSON, &rawCatalog)
	if err != nil {
		log.Println("Warning: Failed to load item catalog:", err)
		return
	}

	// Convert string keys to integers
	for idStr, name := range rawCatalog {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			// Skip invalid IDs
			continue
		}
		ItemCatalog[id] = name
	}

	log.Printf("Loaded %d items into catalog", len(ItemCatalog))
}

// GetItemName returns the name of an item by ID, or empty string if not found
func GetItemName(itemID int) string {
	ItemCatalogLock.RLock()
	defer ItemCatalogLock.RUnlock()

	if name, ok := ItemCatalog[itemID]; ok {
		return name
	}
	return ""
}
