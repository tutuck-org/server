package main

import (
	"log"
	"strings"

	"github.com/chzyer/readline"
)

func startConsole() {
	go func() {
		rl, err := readline.NewEx(&readline.Config{
			Prompt:      "> ",
			HistoryFile: "console.log",
		})
		if err != nil {
			log.Fatal(err)
		}
		defer rl.Close()

		out := ConsoleOutput{}
		out.WriteLine("TuTuck Server Console Started. Type :help for commands.")

		for {
			line, err := rl.Readline()
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
