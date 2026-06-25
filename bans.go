package main

import (
	"encoding/json"
	"os"
	"sync"
)

type BannedStore struct {
	UIDs map[int]bool `json:"uids"`
}

var (
	banned  = BannedStore{UIDs: map[int]bool{}}
	banLock sync.Mutex
)

func loadBans() {
	data, err := os.ReadFile("banned.json")
	if err == nil {
		_ = json.Unmarshal(data, &banned)
	}
}

func saveBans() {
	data, _ := json.MarshalIndent(banned, "", "    ")
	_ = os.WriteFile("banned.json", data, 0644)
}

func banUser(uid int) {
	banLock.Lock()
	defer banLock.Unlock()

	banned.UIDs[uid] = true
	saveBans()
}

func unbanUser(uid int) {
	banLock.Lock()
	defer banLock.Unlock()

	delete(banned.UIDs, uid)
	saveBans()
}

func isBanned(uid int) bool {
	banLock.Lock()
	defer banLock.Unlock()
	return banned.UIDs[uid]
}
