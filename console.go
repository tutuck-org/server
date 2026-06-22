package main

import (
	"fmt"
	"log"
	"strings"

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

		out := ConsoleOutput{}
		out.WriteLine("Console Started. Type :help for commands.")

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

			handleCommand(out, 0, line)
		}
	}()
}

func printToConsole(senderID int, target string, text string) {
	fmt.Print(composeMsg(getName(senderID), target, text))
}
