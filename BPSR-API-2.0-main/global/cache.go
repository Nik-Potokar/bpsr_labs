package global

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Scene monster data
var SceneMonsterList = make(map[uint64]*Monster)
var SceneMonsterListLock = sync.RWMutex{}

var CurrentScene *SceneInfo = nil
var CurrentSceneLock = sync.RWMutex{}

type SceneInfo struct {
	Scene  *SceneData   `json:"scene"`  // Current scene info
	Player *ScenePlayer `json:"player"` // Current scene player info
}
type SceneData struct {
	MapId  uint32 `json:"map_id"`  // Scene ID
	Name   string `json:"name"`    // Scene name
	LineId uint32 `json:"line_id"` // Scene line ID
}
type ScenePlayer struct {
	Id         uint64    `json:"id"`            // Player ID
	FightPoint int32     `json:"fight_point"`   // Combat power
	Name       string    `json:"name"`          // Player nickname
	Level      int32     `json:"level"`         // Player level
	Hp         int64     `json:"hp"`            // Current HP
	MaxHp      int64     `json:"max_hp"`        // Max HP
	Pos        *Position `json:"pos,omitempty"` // Player coordinates
}

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}
type AttackPlayer struct {
	Name           string `json:"name,omitempty"` // Player nickname
	LastAttackTime int64  `json:"-"`              // Last attack time
}
type Monster struct {
	Name            string                   `json:"name,omitempty"`           // Monster name (Chinese)
	NameEN          string                   `json:"name_en,omitempty"`        // Monster name (English)
	Hp              uint64                   `json:"hp"`                       // Current HP
	MaxHp           uint64                   `json:"max_hp,omitempty"`         // Max HP
	Pos             *Position                `json:"pos,omitempty"`            // Monster coordinates
	TemplateId      uint64                   `json:"template_id,omitempty"`    // Template ID
	EntityId        uint64                   `json:"entity_id,omitempty"`      // Current enemy ID
	AttackPlayers   map[uint64]*AttackPlayer `json:"attack_players,omitempty"` // List of attacking players
	UpdateTime      int64                    `json:"-"`                        // Data last update time
}

func ClearAllData() {
	clearMonsterList()
}
func clearMonsterList() {
	SceneMonsterListLock.Lock()
	defer SceneMonsterListLock.Unlock()
	SceneMonsterList = make(map[uint64]*Monster)
}

//func clearScene() {
//	CurrentSceneLock.Lock()
//	defer CurrentSceneLock.Unlock()
//	// Remove player coordinates, scene switching won't change other info, but network interruption causing server re-identification will clear current scene data
//	if CurrentScene != nil && CurrentScene.Player != nil{
//		CurrentScene.Player.Pos = nil
//	}
//}

func FindMonsterId(uuid uint64, callback func(*Monster)) {
	var monster *Monster
	var isNew bool

	// Lock and get or create monster object
	SceneMonsterListLock.Lock()
	if existing, has := SceneMonsterList[uuid]; has {
		monster = existing
		isNew = false
	} else {
		monster = &Monster{
			UpdateTime: time.Now().Unix(),
		}
		SceneMonsterList[uuid] = monster
		isNew = true
	}
	SceneMonsterListLock.Unlock()

	// Call callback outside lock to avoid deadlock and long lock hold
	startTime := time.Now()
	callback(monster)

	// Record performance metrics
	duration := time.Since(startTime).Milliseconds()
	if duration >= 100 {
		log.Println(fmt.Sprintf("%d Abnormal update time: %d ms", uuid, duration))
	}

	// If not a newly created object, update timestamp
	if !isNew {
		SceneMonsterListLock.Lock()
		if currentMonster, exists := SceneMonsterList[uuid]; exists {
			currentMonster.UpdateTime = time.Now().Unix()
		}
		SceneMonsterListLock.Unlock()
	}
}

func UpdateScene(callback func(*SceneInfo)) {
	CurrentSceneLock.Lock()
	defer CurrentSceneLock.Unlock()
	if CurrentScene == nil {
		CurrentScene = &SceneInfo{
			Player: &ScenePlayer{},
			Scene:  &SceneData{},
		}
	}
	callback(CurrentScene)

}

// Market listing data
var MarketListings = make(map[string]*MarketListing)
var MarketListingsLock = sync.RWMutex{}

type MarketListing struct {
	ListingGUID string `json:"listing_guid"`
	ItemID      int    `json:"item_id"`
	ItemName    string `json:"item_name"`
	PriceLuno   int    `json:"price_luno"`
	Quantity    int    `json:"quantity"`
	BindFlag    bool   `json:"bind_flag"`
	ListedAt    int64  `json:"listed_at"`   // Unix timestamp when listed
	CapturedAt  int64  `json:"captured_at"` // When we captured this
	UnitPrice   int    `json:"unit_price"`  // Price per item
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

// AddMarketListing adds a new market listing and records it in price history
func AddMarketListing(listing *MarketListing) {
	MarketListingsLock.Lock()
	MarketListings[listing.ListingGUID] = listing
	MarketListingsLock.Unlock()

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

// CleanOldListings removes listings older than maxAge seconds
func CleanOldListings(maxAge int64) int {
	MarketListingsLock.Lock()
	defer MarketListingsLock.Unlock()

	now := time.Now().Unix()
	removed := 0
	for guid, listing := range MarketListings {
		if now-listing.CapturedAt > maxAge {
			delete(MarketListings, guid)
			removed++
		}
	}
	return removed
}

// ClearMarketData clears all market listings and price history
func ClearMarketData() {
	MarketListingsLock.Lock()
	MarketListings = make(map[string]*MarketListing)
	MarketListingsLock.Unlock()

	PriceHistoryLock.Lock()
	PriceHistory = make(map[int][]PriceRecord)
	PriceHistoryLock.Unlock()
}
