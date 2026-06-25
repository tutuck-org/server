package main

import "time"

const (
	ServerID    = 0
	BroadcastID = -1
)

var Version string

var ServerInfo struct {
	Fingerprint string
	StartTime   time.Time
}

func initServerUser() {
	userLock.Lock()
	defer userLock.Unlock()

	if findUser(ServerID) != nil {
		return
	}

	userStore.Users = append(userStore.Users, User{
		ID:   ServerID,
		Name: "Server",
		Key:  "",
	})

	saveUsers()
}
