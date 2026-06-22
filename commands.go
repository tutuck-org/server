package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func viewOnline(out Output) {
	clLock.Lock()
	defer clLock.Unlock()
	if len(clients) == 0 {
		out.WriteLine("Error: no users online")
		return
	}
	var uids []string
	for uid := range clients {
		uids = append(uids, fmt.Sprintf("%s", getName(uid)))
	}
	out.WriteLine("Online users: " + strings.Join(uids, ", "))
}

func checkWho(out Output, line string) {
	target := strings.TrimPrefix(line, "@")
	userLock.Lock()
	tUser := findUser(target)
	userLock.Unlock()

	if tUser == nil {
		out.WriteLine("Error: User not found")
		return
	}

	response := fmt.Sprintf("%s (%d), pubkey:\n%s", tUser.Name, tUser.ID, tUser.Key)
	out.WriteLine(response)
}

func handleCommand(out Output, uid int, msg string) {
	fields := strings.Fields(msg)

	if strings.HasPrefix(msg, ":") {
		cmd := fields[0]
		switch cmd {
		case ":help":
			helpText := `
TuTuck Help
===========

Chatting:
  just <message>     → send to everyone
  :dm <uid|name>     → start private chat
  :dm off            → exit DM
  @<uid|name> <msg>  → message user
  @* or @everyone    → broadcast to all
  @0 or @server      → message server
  :me <action>       → describe your action

  :info              → check server info
  :online or :ls     → see online users
  :who <uid|name>    → get user info
  :name [change]     → show or change your username
`
			out.WriteLine(helpText)
		case ":dm":
			if len(fields) < 2 {
				out.WriteLine("Usage: :dm <uid|name|off>")
				return
			}
			arg := fields[1]
			if arg == "off" || arg == "exit" {
				clearActiveDM(uid)
				out.WriteLine("Exited DM mode")
				return
			}
			userLock.Lock()
			tUser := findUser(arg)
			userLock.Unlock()
			if tUser == nil {
				out.WriteLine("Error: user not found")
				return
			}
			if tUser.ID == uid {
				out.WriteLine("Error: cannot DM yourself")
				return
			}
			clLock.Lock()
			_, online := clients[tUser.ID]
			clLock.Unlock()
			if !online {
				out.WriteLine("Error: user is not online")
				return
			}
			setActiveDM(uid, tUser.ID)
			out.WriteLine(fmt.Sprintf("You entered DM with %s", tUser.Name))
		case ":info", ":about":
			dmLog := "disabled"
			if cfg.LogDMs {
				dmLog = "enabled (!)"
			}

			clLock.Lock()
			onlineCount := len(clients)
			clLock.Unlock()
			// TODO: use admin's username instead of @Server
			infoText := fmt.Sprintf(`
TuTuck Server Info
==================
Version: %s
Uptime: %s

DM logging %s
Online: %d/%d

Report any issues to @Server

Fingerprint:
  %s
`, Version, time.Since(ServerInfo.StartTime).Round(time.Second), dmLog, onlineCount, cfg.MaxClients, ServerInfo.Fingerprint)
			out.WriteLine(infoText)
		case ":online", ":ls":
			viewOnline(out)
		case ":me":
			if len(fields) < 2 {
				out.WriteLine("Usage: :me <action>")
				return
			}
			broadcastAction(uid, strings.Join(fields[1:], " "))
		case ":who":
			checkWho(out, strings.Join(fields[1:], " "))
		case ":name":
			if len(fields) == 1 {
				out.WriteLine(fmt.Sprintf("Your current name is: %s", getName(uid)))
				return
			}
			if len(fields) >= 2 && strings.ToLower(fields[1]) == "change" {
				if ch, ok := out.(ChannelOutput); ok {
					changeName(ch.ch, uid, false)
				} else {
					out.WriteLine("Name change available only for connected users")
				}
				return
			}
			out.WriteLine("Usage:\n  :name        → show your name\n  :name change → change your username")
		case ":stop":
			if uid == ServerID {
				os.Exit(0)
			}
			return
		// TODO: implement client disconnect via command
		case ":leave", ":quit", ":exit":
			/*if ch, ok := out.(ChannelOutput); ok {
				ch.ch.Close()
			}*/
			return
		default:
			out.WriteLine("Unknown command")
		}
		return
	}

	if strings.HasPrefix(msg, "@") {
		if len(fields) < 2 {
			out.WriteLine("Usage: @<uid|name> <message>")
			return
		}
		target := strings.TrimPrefix(fields[0], "@")
		text := strings.Join(fields[1:], " ")
		sendMessage(out, uid, target, text)
		return
	}

	if targetID, ok := getActiveDM(uid); ok {
		clLock.Lock()
		recv := clients[targetID]
		clLock.Unlock()
		if recv == nil {
			clearActiveDM(uid)
			out.WriteLine("Error: target went offline, exited DM mode")
			return
		}
		sendToUser(out, uid, fmt.Sprintf("%d", targetID), msg)
	} else {
		sendToAll(uid, msg)
	}
}
