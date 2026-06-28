package main

import (
	"fmt"
	"log"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

func startConsole() {
	go func() {
		rl, err := readline.NewEx(&readline.Config{
			Prompt:          "> ",
			HistoryFile:     "console.log",
			InterruptPrompt: "^C",
			EOFPrompt:       "exit",
		})
		if err != nil {
			log.Fatal(err)
		}
		defer rl.Close()

		fmt.Println("Console Started. Type /help for commands.")

		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				shutdownServer("Killed")
				return
			}
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			handleConsole(line)
		}
	}()
}

func shutdownServer(reason string) {
	clLock.Lock()
	clientsCopy := maps.Clone(clients)
	clLock.Unlock()

	for _, ch := range clientsCopy {
		sendErrPacket(ch, "Server is shutting down! Reason: %s", reason)
	}

	time.Sleep(100 * time.Millisecond)

	for _, ch := range clientsCopy {
		ch.Close()
	}

	fmt.Printf("Stopping! Server was up for %s. Disconnecting %d clients\n", time.Since(ServerInfo.StartTime).Round(time.Second), len(clientsCopy))
	os.Exit(0)
}

func handleConsole(line string) {
	fields := strings.Fields(line)

	if strings.HasPrefix(line, "/") {
		cmd := strings.TrimPrefix(fields[0], "/")
		switch cmd {
		case "help":
			helpText := `
TuTuck Server Help
==================

  /info              → check server info 
  /online or /ls     → see online users
  /who <uid|name>    → get user info
  /ban <uid|name>
  /unban <uid|name>
  /stop [reason]
`
			fmt.Println(helpText)
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
Admin: %s

Fingerprint:
  %s
`, Version, time.Since(ServerInfo.StartTime).Round(time.Second), dmLog, onlineCount, cfg.MaxClients, cfg.Admin, ServerInfo.Fingerprint)
			fmt.Println(infoText)
		case "online", "ls":
			clLock.Lock()
			defer clLock.Unlock()
			if len(clients) == 0 {
				fmt.Println("Error: no users online")
				return
			}
			var uNames []string
			for uid := range clients {
				uNames = append(uNames, fmt.Sprintf("%s", getName(uid)))
			}
			fmt.Printf("Online users: \n%s\n", strings.Join(uNames, ", "))
		case "who":
			if len(fields) < 2 {
				fmt.Println("Usage: /who <username or id>")
				return
			}
			target := strings.TrimPrefix(fields[1], "@")
			userLock.Lock()
			tUser := findUser(target)
			userLock.Unlock()

			if tUser == nil {
				fmt.Println("Error: User not found")
				return
			}

			fmt.Printf("%s (%d), pubkey:\n%s\n", tUser.Name, tUser.ID, tUser.Key)
		case "kick":
			// TODO: fix client auto-reconnect
			if len(fields) < 2 {
				fmt.Println("Usage: /kick <uid|name>")
				return
			}

			target := strings.TrimPrefix(fields[1], "@")
			userLock.Lock()
			tUser := findUser(target)
			userLock.Unlock()

			if tUser == nil {
				fmt.Println("Error: user not found")
				return
			}

			if tUser.ID == ServerID {
				fmt.Println("Error: not kicking server today")
				return
			}

			clLock.Lock()
			if ch, ok := clients[tUser.ID]; ok {
				ch.Close()
			}
			clLock.Unlock()

			printToConsole(tUser.ID, "was kicked")
		case "ban":
			if len(fields) < 2 {
				fmt.Println("Usage: /ban <uid|name>")
				return
			}

			target := strings.TrimPrefix(fields[1], "@")
			userLock.Lock()
			tUser := findUser(target)
			userLock.Unlock()

			if tUser == nil {
				fmt.Println("Error: user not found")
				return
			}

			if tUser.ID == ServerID {
				fmt.Println("Error: not banning server today")
				return
			}

			banUser(tUser.ID)

			clLock.Lock()
			if ch, ok := clients[tUser.ID]; ok {
				ch.Close()
			}
			clLock.Unlock()

			deliverMessage(ServerID, BroadcastID, ScopeGlobal, fmt.Sprintf("%s got banned", tUser.Name))
			fmt.Printf("Banned %s (%d)\n", tUser.Name, tUser.ID)
		case "unban":
			if len(fields) < 2 {
				fmt.Println("Usage: /unban <uid|name>")
				return
			}

			target := strings.TrimPrefix(fields[1], "@")
			userLock.Lock()
			tUser := findUser(target)
			userLock.Unlock()

			if tUser == nil {
				fmt.Println("Error: user not found")
				return
			}

			unbanUser(tUser.ID)

			deliverMessage(ServerID, BroadcastID, ScopeGlobal, fmt.Sprintf("%s got unbanned", tUser.Name))
			fmt.Printf("Unbanned %s (%d)\n", tUser.Name, tUser.ID)
		case "stop":
			var reason string
			if len(fields) < 2 {
				reason = "Server is shutting down."
			} else {
				reason = strings.Join(fields[1:], " ")
			}

			shutdownServer(reason)
		default:
			fmt.Println("Unknown command")
		}
		return
	}

	if strings.HasPrefix(line, "@") {
		if len(fields) < 2 {
			fmt.Println("Usage: @<uid|name> <message>")
			return
		}
		target := strings.TrimPrefix(fields[0], "@")
		tUser := findUser(target)

		if tUser == nil {
			fmt.Println("User not found")
			return
		}

		text := strings.Join(fields[1:], " ")
		deliverMessage(ServerID, tUser.ID, ScopeDM, text)
		fmt.Printf("Messaged @%s (%d)\n", tUser.Name, tUser.ID)
		return
	}
	broadcastMsg(ServerID, line)
}

func printToConsole(uid int, format string, args ...any) {
	fmt.Printf("%s (%d) %s\n", getName(uid), uid, fmt.Sprintf(format, args...))
}
