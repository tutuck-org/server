package main

import (
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Color struct {
	Name string
	Hex  string
}

var Colors = []Color{
	{"Red", "#FF5555"},
	{"Dark Red", "#CC0000"},
	{"Orange", "#FF8800"},
	{"Amber", "#FFBF00"},
	{"Yellow", "#FFFF55"},
	{"Lime", "#AAFF00"},
	{"Green", "#55FF55"},
	{"Emerald", "#00CC66"},
	{"Mint", "#66FFCC"},
	{"Teal", "#00AAAA"},
	{"Cyan", "#55FFFF"},
	{"Sky", "#66CCFF"},
	{"Blue", "#5599FF"},
	{"Royal Blue", "#3366FF"},
	{"Navy", "#0033AA"},
	{"Purple", "#AA55FF"},
	{"Violet", "#8844FF"},
	{"Magenta", "#FF55FF"},
	{"Pink", "#FF77CC"},
	{"Rose", "#FF6699"},
	{"Coral", "#FF7F50"},
	{"Salmon", "#FA8072"},
	{"Brown", "#8B4513"},
	{"Chocolate", "#D2691E"},
	{"Gold", "#FFD700"},
	{"Silver", "#C0C0C0"},
	{"Gray", "#808080"},
	{"White", "#FFFFFF"},
	{"Turquoise", "#40E0D0"},
	{"Indigo", "#4B0082"},
	{"Lavender", "#B57EDC"},
	{"Crimson", "#DC143C"},
}

func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")

	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)

	return int(r), int(g), int(b)
}

func chooseColor(ch ssh.Channel, uid int) {
	for i, c := range Colors {
		sendColorPacket(ch, i+1, c)
	}

	sendSysPacket(ch, "Changing the color of your name. Enter number or color name: ")

	for {
		buf := make([]byte, 64)
		n, err := ch.Read(buf)
		if err != nil || n == 0 {
			return
		}

		input := strings.TrimSpace(string(buf[:n]))

		var choice *Color

		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= len(Colors) {
				choice = &Colors[num-1]
			}
		}

		if choice == nil {
			for i := range Colors {
				if strings.EqualFold(Colors[i].Name, input) {
					choice = &Colors[i]
					break
				}
			}
		}

		if choice == nil {
			sendSysPacket(ch, "Invalid color. Try again.")
			continue
		}

		userLock.Lock()
		if user := findUser(uid); user != nil {
			user.Color = choice.Hex
		}
		userLock.Unlock()

		saveUsers()
		sendSysPacket(ch, "Your new color is %s (%s)", choice.Name, choice.Hex)
		return
	}
}

func getColor(uid int) string {
	userLock.Lock()
	u := findUser(uid)
	userLock.Unlock()
	if u != nil && u.Color != "" {
		return u.Color
	}
	return "#FFFFFF"
}
