package main

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type Message struct {
	From string
	To   string
	Time string
	Text string
}

func composeMsg(from, to, text string) Message {
	return Message{
		From: from,
		To:   to,
		Time: time.Now().Format("15:04"),
		Text: text}
}

func (m Message) String() string {
	return fmt.Sprintf(
		"%s -> %s | %s\n: %s\n",
		m.From, m.To, m.Time, m.Text)
}

func sendMsg(out any, msg Message) {
	switch ch := out.(type) {
	case ssh.Channel:
		ch.Write([]byte(msg.String()))
	case Output:
		ch.WriteLine(msg.String())
	default:
		fmt.Printf("Unknown output type: %T\n%s", out, msg.String())
	}
}

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
	printToConsole(senderID, "Server", text)
	sendMsg(out, composeMsg(getName(senderID), "Server", text))
}

func sendToAll(senderID int, text string) {
	clLock.Lock()
	defer clLock.Unlock()
	for _, c := range clients {
		sendMsg(c, composeMsg(getName(senderID), "All", text))
	}
	printToConsole(senderID, "All", text)
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

	sendMsg(out, composeMsg(getName(senderID), getName(receiverID), text))
	sendMsg(recv, composeMsg(getName(senderID), getName(receiverID), text))
	if cfg.LogDMs {
		printToConsole(senderID, getName(receiverID), text)
	}
}
