package main

import (
	"fmt"
	"time"
)

func deliverMessage(from, to int, scope ScopeType, content string) {
	clLock.Lock()
	defer clLock.Unlock()

	switch to {
	case BroadcastID:
		for id, ch := range clients {
			direction := "in"
			if id == from {
				direction = "out"
			}
			sendMsgPacket(ch, from, BroadcastID, scope, direction, content)
		}
		return
	case ServerID:
		fmt.Printf("%s (%d) | %s \n: %s\n", getName(from), from, time.Now().Format("15:04"), content)
	}

	toCh := clients[to]
	fromCh := clients[from]

	if toCh != nil {
		sendMsgPacket(toCh, from, to, scope, "in", content)
	}

	if fromCh != nil && from != ServerID {
		sendMsgPacket(fromCh, from, to, scope, "out", content)
	}
}
