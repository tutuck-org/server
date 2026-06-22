package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var reservedNames = map[string]struct{}{
	"server":    {},
	"host":      {},
	"console":   {},
	"all":       {},
	"yourself":  {},
	"everybody": {},
	"everyone":  {},
	"you":       {},
}

func isMessageTooLong(msg string) bool {
	return len(msg) > 2048
}

func getName(uid int) string {
	if uid == 0 {
		return "Server"
	}
	userLock.Lock()
	u := findUser(uid)
	userLock.Unlock()
	if u != nil && u.Name != "" {
		return u.Name
	}
	return fmt.Sprintf("User %d", uid)
}

func findUser(uidOrName any) *User {
	switch v := uidOrName.(type) {
	case int:
		for i := range userStore.Users {
			if userStore.Users[i].ID == v {
				return &userStore.Users[i]
			}
		}
	case string:
		if uid, err := strconv.Atoi(v); err == nil {
			for i := range userStore.Users {
				if userStore.Users[i].ID == uid {
					return &userStore.Users[i]
				}
			}
		} else {
			lower := strings.ToLower(v)
			for i := range userStore.Users {
				if strings.ToLower(userStore.Users[i].Name) == lower {
					return &userStore.Users[i]
				}
			}
		}
	}
	return nil
}

func sendMsg(to any, from, toName, text string) {
	timeStr := time.Now().Format("15:04")
	msg := fmt.Sprintf("%s -> %s | %s\n: %s\n", from, toName, timeStr, text)

	switch ch := to.(type) {
	case ssh.Channel:
		ch.Write([]byte(msg))
	case Output:
		ch.WriteLine(msg)
	default:
		fmt.Printf("Unknown output type: %T\n%s", to, msg)
	}
}

func printMsg(from, to, text string) {
	timeStr := time.Now().Format("15:04")

	msg := fmt.Sprintf("%s -> %s | %s \n: %s \n",
		from, to,
		timeStr,
		text,
	)

	fmt.Printf("%s", msg)
}

func changeName(ch ssh.Channel, uid int, firstTime bool) {
	for {
		if firstTime {
			ch.Write([]byte("Welcome to TuTuck! Please set a username: "))
		} else {
			ch.Write([]byte("Enter a new username: "))
		}

		buf := make([]byte, 256)
		n, err := ch.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		input := strings.TrimSpace(string(buf[:n]))

		if input == "" {
			ch.Write([]byte("Username cannot be empty, try again.\n"))
			continue
		}

		valid := true
		for _, r := range input {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				valid = false
				break
			}
		}
		if !valid {
			ch.Write([]byte("Username can only contain letters a-z or A-Z, try again.\n"))
			continue
		}

		userLock.Lock()
		duplicate := false
		for _, u := range userStore.Users {
			if strings.EqualFold(u.Name, input) {
				duplicate = true
				break
			}
		}
		if duplicate {
			userLock.Unlock()
			ch.Write([]byte("This username is already taken, choose another.\n"))
			continue
		}

		if _, ok := reservedNames[strings.ToLower(input)]; ok {
			userLock.Unlock()
			ch.Write([]byte("This username is reserved, choose another.\n"))
			continue
		}

		var oldName string
		if !firstTime {
			if u := findUser(uid); u != nil {
				oldName = u.Name
			}
		}

		user := findUser(uid)
		if user != nil {
			user.Name = input
		}

		userLock.Unlock()
		saveUsers()

		if firstTime {
			ch.Write([]byte(fmt.Sprintf("You registered as %s!\n", input)))
		} else {
			ch.Write([]byte(fmt.Sprintf("You changed your name from %s to %s!\n", oldName, input)))
		}
		break
	}
}
