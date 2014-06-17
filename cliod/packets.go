package cliod

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
	ID uint16
}

type EncryptedPacket struct {
	Payload []byte
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
	T1, T2    int32
}

func (p *PacketHello) A() {}
func (p *PacketAuth) A()  {}
