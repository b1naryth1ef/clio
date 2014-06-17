package cliod

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"github.com/b1naryth1ef/json"
// 	"log"
// )

type PacketAuthPayload struct {
	T1, T2 int32
}

type Packet interface {
	A()
}

type BasePacket struct {
	ID uint16
}

type PacketHello struct {
	ID          uint16
	PublicKey   []byte
	NetworkHash string
	Token       int32
}

type PacketAuth struct {
	ID        uint16
	PublicKey []byte
	Payload   []byte
}

func (p *PacketHello) A() {}
func (p *PacketAuth) A()  {}
