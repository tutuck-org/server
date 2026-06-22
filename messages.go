package main

import (
	"strconv"
)

func sendMessage(out Output, senderID int, target string, text string) {
	switch target {
	case "0", "server", "host", "console":
		sendToServer(out, senderID, text)
	case "*", "all", "everyone", "everybody":
		sendToAll(senderID, text)
	default:
		sendToUser(out, senderID, target, text)
	}
}

func sendToServer(out Output, senderID int, text string) {
	printMsg(getName(senderID), "Server", text)
	sendMsg(out, "You", "Server", text)
}

func sendToAll(senderID int, text string) {
	clLock.Lock()
	defer clLock.Unlock()
	for id, c := range clients {
		if id == senderID {
			sendMsg(c, "You", "All", text)
		} else {
			sendMsg(c, getName(senderID), "All", text)
		}
	}
	printMsg(getName(senderID), "All", text)
}

func sendToUser(out Output, senderID int, target string, text string) {
	var receiverID int
	if id, err := strconv.Atoi(target); err == nil {
		receiverID = id
	} else {
		userLock.Lock()
		tUser := findUser(target)
		userLock.Unlock()
		if tUser == nil {
			out.WriteLine("Error: user not found")
			return
		}
		receiverID = tUser.ID
	}

	clLock.Lock()
	recv := clients[receiverID]
	clLock.Unlock()

	if recv == nil {
		out.WriteLine("Error: receiver is not connected")
		return
	}

	if receiverID == senderID {
		sendMsg(out, "You", "Yourself", text)
		return
	}

	sendMsg(out, "You", getName(receiverID), text)
	sendMsg(recv, getName(senderID), "You", text)
	printMsg(getName(senderID), getName(receiverID), text)
}
