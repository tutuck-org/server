package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func viewOnline(ch ssh.Channel) {
	clLock.Lock()
	defer clLock.Unlock()
	if len(clients) == 0 {
		sendErrPacket(ch, "Error: no users online")
		return
	}
	var uNames []string
	for uid := range clients {
		uNames = append(uNames, fmt.Sprintf("%s", getName(uid)))
	}
	sendSysPacket(ch, "Online users: %s", strings.Join(uNames, ", "))
}

func checkWho(ch ssh.Channel, uid int, line string) {
	target := strings.TrimPrefix(line, "@")
	userLock.Lock()
	tUser := findUser(target)
	userLock.Unlock()

	if tUser == nil {
		sendErrPacket(ch, "Error: User not found")
		return
	}

	if uid == tUser.ID {
		sendSysPacket(ch, "You are %s (%d), your pubkey:\n%s", tUser.Name, tUser.ID, tUser.Key)
		return
	}
	sendSysPacket(ch, "%s (%d), pubkey:\n%s", tUser.Name, tUser.ID, tUser.Key)
}

func handleCommand(ch ssh.Channel, uid int, msg string) {
	fields := strings.Fields(msg)

	if strings.HasPrefix(msg, "/") {
		cmd := strings.TrimPrefix(fields[0], "/")
		switch cmd {
		case "help":
			helpText := `
TuTuck Help
===========
  just <message>     → send to everyone
  /dm <uid|name>     → start private chat
  /dm off            → exit DM
  @<uid|name> <msg>  → message user
  @server            → message server
  /me <action>       → describe your action

  /info              → check server info
  /online or :ls     → see online users
  /whoami            → get to know who you are
  /who <uid|name>    → get user info
  /customize         → change name or color
`
			sendSysPacket(ch, "%s", helpText)
		case "dm":
			if len(fields) < 2 {
				sendSysPacket(ch, "Usage: /dm <uid|name|off>")
				return
			}
			arg := fields[1]
			if arg == "off" || arg == "exit" {
				clearActiveDM(uid)
				sendSysPacket(ch, "Exited DM mode")
				return
			}
			userLock.Lock()
			tUser := findUser(arg)
			userLock.Unlock()
			if tUser == nil {
				sendErrPacket(ch, "Error: user not found")
				return
			}
			if tUser.ID == uid {
				sendErrPacket(ch, "Error: cannot DM yourself")
				return
			}
			clLock.Lock()
			_, online := clients[tUser.ID]
			clLock.Unlock()
			if !online {
				sendErrPacket(ch, "Error: user is not online")
				return
			}
			setActiveDM(uid, tUser.ID)
			sendSysPacket(ch, "You entered DM with %s", tUser.Name)
		case "info", "about":
			dmLog := "disabled"
			if cfg.LogDMs {
				dmLog = "enabled (!)"
			}

			clLock.Lock()
			onlineCount := len(clients)
			clLock.Unlock()
			infoText := fmt.Sprintf(`
TuTuck Server Info
==================
Version: %s
Uptime: %s

DM logging %s
Online: %d/%d

Report any issues to @%s

Fingerprint:
  %s
`, Version, time.Since(ServerInfo.StartTime).Round(time.Second), dmLog, onlineCount, cfg.MaxClients, cfg.Admin, ServerInfo.Fingerprint)
			sendSysPacket(ch, "%s", infoText)
		case "online", "ls":
			viewOnline(ch)
		case "me":
			if len(fields) < 2 {
				sendSysPacket(ch, "Usage: /me <action>")
				return
			}
			broadcastAction(uid, strings.Join(fields[1:], " "))
		case "whoami":
			checkWho(ch, uid, getName(uid))
		case "who":
			if len(fields) < 2 {
				sendErrPacket(ch, "Usage: /who <uid|name>")
				return
			}
			checkWho(ch, uid, strings.Join(fields[1:], " "))
		case "customize", "custom":
			if len(fields) < 2 {
				sendSysPacket(ch, `
Usage:
  /customize name       → change username
  /customize color      → change name color
`)
				return
			}

			switch strings.ToLower(fields[1]) {
			case "name":
				sendSysPacket(ch, "Your current name is: %s", getName(uid))
				changeName(ch, uid, false)
				return
			case "color":
				chooseColor(ch, uid)
				return
			}
		// TODO: fix client disconnect (client has autoreconnect)
		case "leave", "quit", "exit":
			ch.Close()
			return
		default:
			sendErrPacket(ch, "Unknown command")
		}
		return
	}

	if strings.HasPrefix(msg, "@") {
		if len(fields) < 2 {
			sendSysPacket(ch, "Usage: @<uid|name> <message>")
			return
		}
		target := strings.TrimPrefix(fields[0], "@")
		tUser := findUser(target)

		if tUser == nil {
			sendErrPacket(ch, "User not found")
			return
		}

		text := strings.Join(fields[1:], " ")
		deliverMessage(uid, tUser.ID, ScopeDM, text)
		return
	}

	if targetID, ok := getActiveDM(uid); ok {
		clLock.Lock()
		recv := clients[targetID]
		clLock.Unlock()
		if recv == nil {
			clearActiveDM(uid)
			sendErrPacket(ch, "Error: target went offline, exited DM mode")
			return
		}
		deliverMessage(uid, targetID, ScopeDM, msg)
	} else {
		broadcastMsg(uid, msg)
	}
}
