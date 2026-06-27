package main

import (
	"fmt"
	"maps"
	"time"
)

func broadcastJoin(uid int, firstTime bool) {
	clLock.Lock()
	clientsCopy := maps.Clone(clients)
	clLock.Unlock()

	for id, ch := range clientsCopy {
		if id == uid {
			sendSysPacket(ch, "You (%s) joined at %s", getName(uid), time.Now().Format("15:04"))

			if firstTime {
				sendSysPacket(ch, "Tip: use :help to see available commands")
			}
		} else {
			sendSysPacket(ch, "%s joined at %s", getName(uid), time.Now().Format("15:04"))
		}
	}

	printToConsole(uid, "joined at %s", time.Now().Format("15:04"))
}

func broadcastLeave(uid int) {
	clLock.Lock()
	clientsCopy := maps.Clone(clients)
	clLock.Unlock()

	for _, ch := range clientsCopy {
		sendSysPacket(ch, "%s left at %s\n", getName(uid), time.Now().Format("15:04"))
	}

	printToConsole(uid, "left at %s", time.Now().Format("15:04"))
	clearActiveDMByTarget(uid)
}

func broadcastAction(uid int, action string) {
	clLock.Lock()
	clientsCopy := maps.Clone(clients)
	clLock.Unlock()

	for id, ch := range clientsCopy {
		if id == uid {
			sendSysPacket(ch, "You (%s) %s", getName(uid), action)
		} else {
			sendSysPacket(ch, "%s %s", getName(uid), action)
		}
	}
	fmt.Printf("%s (%d) %s\n", getName(uid), uid, action)
}

func broadcastMsg(from int, content string) {
	deliverMessage(from, BroadcastID, ScopeGlobal, content)

	if cfg.EchoMsgs {
		fmt.Printf("%s (%d) | %s\n: %s\n", getName(from), from, time.Now().Format("15:04"), content)
	}
}
