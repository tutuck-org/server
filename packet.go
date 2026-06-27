package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

type (
	PacketType string
	ScopeType  string
)

const (
	TypeMessage  PacketType = "message"
	TypeSystem   PacketType = "system"
	TypeError    PacketType = "error"
	TypeIdentity PacketType = "identity"
	TypeColor    PacketType = "color"

	ScopeDM     ScopeType = "dm"
	ScopeGlobal ScopeType = "global"
)

type Packet struct {
	// Identity
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Type PacketType `json:"type"`

	FromID    int    `json:"fromid,omitempty"`
	From      string `json:"from,omitempty"`
	FromColor string `json:"fromcolor,omitempty"`
	ToID      int    `json:"toid,omitempty"`
	To        string `json:"to,omitempty"`
	ToColor   string `json:"tocolor,omitempty"`

	Scope     ScopeType `json:"scope,omitempty"`
	Direction string    `json:"direction,omitempty"`
	Content   string    `json:"content,omitempty"`
	Time      string    `json:"time,omitempty"`

	Num       int    `json:"num,omitempty"`
	ColorName string `json:"colorname,omitempty"`
	ColorHex  string `json:"colorhex,omitempty"`
}

func sendPacket(w io.Writer, p Packet) error {
	return json.NewEncoder(w).Encode(p)
}

func sendMsgPacket(ch ssh.Channel, from, to int, scope ScopeType, direction, content string) {
	sendPacket(ch, Packet{
		Type:      TypeMessage,
		FromID:    from,
		From:      getName(from),
		FromColor: getColor(from),
		ToID:      to,
		To:        getName(to),
		ToColor:   getColor(to),
		Scope:     scope,
		Direction: direction,
		Content:   content,
		Time:      time.Now().Format("15:04"),
	})
}

func sendSysPacket(ch ssh.Channel, format string, args ...any) {
	sendPacket(ch, Packet{
		Type:    TypeSystem,
		Content: fmt.Sprintf(format, args...),
	})
}

func sendErrPacket(ch ssh.Channel, format string, args ...any) {
	sendPacket(ch, Packet{
		Type:    TypeError,
		Content: fmt.Sprintf(format, args...),
	})
}

func sendColorPacket(ch ssh.Channel, i int, c Color) {
	sendPacket(ch, Packet{
		Type:      TypeColor,
		Num:       i,
		ColorName: c.Name,
		ColorHex:  c.Hex,
	})
}
