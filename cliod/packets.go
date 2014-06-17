package cliod

import "time"

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"github.com/b1naryth1ef/json"
// 	"log"
// )

type Packet interface {
	A()
}

type BasePacket struct {
	ID   uint16
	Time time.Time
}

func NewBasePacket(ID uint16) BasePacket {
	return BasePacket{
		ID:   ID,
		Time: time.Now(),
	}
}

type EncryptedPacket struct {
	Payload []byte
}

type PacketHello struct {
	BasePacket
	PublicKey   []byte
	NetworkHash string
	Token       int32
}

type PacketAuth struct {
	BasePacket
	PublicKey []byte
	T1        int32
}

func (p *PacketHello) A() {}
func (p *PacketAuth) A()  {}
