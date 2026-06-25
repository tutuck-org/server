package main

import (
	"fmt"
	"strconv"
	"strings"

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
	if uid == ServerID {
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

func changeName(ch ssh.Channel, uid int, firstTime bool) {
	if firstTime {
		sendSysPacket(ch, "Welcome to TuTuck! Please set a username: ")
	} else {
		sendSysPacket(ch, "Enter a new username: ")
	}

	for {
		buf := make([]byte, 256)
		n, err := ch.Read(buf)
		if err != nil || n == 0 {
			return
		}
		input := strings.TrimSpace(string(buf[:n]))

		if input == "" {
			sendSysPacket(ch, "Username cannot be empty, try again.\n")
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
			sendSysPacket(ch, "Username can only contain letters a-z or A-Z, try again.\n")
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
			sendSysPacket(ch, "This username is already taken, choose another.\n")
			continue
		}

		if _, ok := reservedNames[strings.ToLower(input)]; ok {
			userLock.Unlock()
			sendSysPacket(ch, "This username is reserved, choose another.\n")
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
			sendSysPacket(ch, "You registered as %s!\n", input)
		} else {
			sendSysPacket(ch, "You changed your name from %s to %s!\n", oldName, input)
		}
		break
	}
}
