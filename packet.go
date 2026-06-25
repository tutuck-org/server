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

	ScopeDM     ScopeType = "dm"
	ScopeGlobal ScopeType = "global"
)

type Packet struct {
	// Identity
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Type PacketType `json:"type"`

	FromID    int       `json:"from,omitempty"`
	ToID      int       `json:"to,omitempty"`
	Scope     ScopeType `json:"scope,omitempty"`
	Direction string    `json:"direction,omitempty"`
	Content   string    `json:"content,omitempty"`
	Time      string    `json:"time,omitempty"`
}

func sendPacket(w io.Writer, p Packet) error {
	return json.NewEncoder(w).Encode(p)
}

func sendMsgPacket(ch ssh.Channel, from, to int, scope ScopeType, direction, content string) {
	sendPacket(ch, Packet{
		Type:      TypeMessage,
		FromID:    from,
		ToID:      to,
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
