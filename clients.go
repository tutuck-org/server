package main

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

var (
	clients = make(map[int]ssh.Channel)
	clLock  sync.Mutex
)

func handleClient(ch ssh.Channel, uid int) {
	defer func() {
		clLock.Lock()
		if c, ok := clients[uid]; ok {
			delete(clients, uid)
			_ = c.Close()
		}
		clLock.Unlock()
		broadcastLeave(uid)
	}()

	userLock.Lock()
	user := findUser(uid)
	userLock.Unlock()

	firstTime := false
	if user.Name == "" {
		firstTime = true
	}

	if firstTime {
		changeName(ch, uid, firstTime)
	}

	userLock.Lock()
	user = findUser(uid)
	userLock.Unlock()

	sendPacket(ch, Packet{
		Type: TypeIdentity,
		ID:   uid,
		Name: user.Name,
	})

	broadcastJoin(uid, firstTime)
	viewOnline(ch)

	buf := make([]byte, 2048)

	for {
		n, err := ch.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Read error:", err)
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))
		if msg == "" || isMessageTooLong(msg) {
			if isMessageTooLong(msg) {
				sendErrPacket(ch, "Error: message is too long (max 2048 chars)")
			}
			continue
		}

		handleCommand(ch, uid, msg)
	}
}
