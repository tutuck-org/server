package main

import (
	"fmt"
	"log"
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

		fmt.Println("Console Started. Type :help for commands.")

		for {
			line, err := rl.Readline()
			if err == readline.ErrInterrupt {
				fmt.Println("Press ^C again to stop server")
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

func handleConsole(line string) {
	fields := strings.Fields(line)

	if strings.HasPrefix(line, ":") {
		cmd := fields[0]
		switch cmd {
		case ":help":
			helpText := `
TuTuck Server Help
==================

  :info              → check server info 
  :online or :ls     → see online users
  :who <uid|name>    → get user info
`
			fmt.Println(helpText)
		case ":info", ":about":
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
		case ":online", ":ls":
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
		case ":who":
			if len(fields) < 2 {
				fmt.Println("Usage: :who <username or id>")
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
		case ":stop":
			os.Exit(0)
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
