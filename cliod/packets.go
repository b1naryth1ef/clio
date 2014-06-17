package cliod

import "time"

type Packet interface {
	SetID()
}

type BasePacket struct {
	ID   uint16
	Time time.Time
}

type PacketPeer struct {
	ID [20]byte
	IP string
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

type PacketPing struct {
	BasePacket
	Token int
}

type PacketPeerListRequest struct {
	BasePacket
	MaxPeers int32
}

type PacketPeerListSync struct {
	BasePacket
	Peers []PacketPeer
}

type PacketNewCrate struct {
	BasePacket
	Crate Crate
}

type PacketSyncCrates struct {
	BasePacket
	Crates []Crate
}

type PacketSearchCrates struct {
	BasePacket
	CrateID    string
	Tags       []string
	BeforeTime time.Time
	AfterTime  time.Time
}

// Interfacers
func (p *PacketHello) SetID() {
	p.BasePacket = NewBasePacket(1)
}

func (p *PacketAuth) SetID() {
	p.BasePacket = NewBasePacket(2)
}

func (p *PacketPing) SetID() {
	p.BasePacket = NewBasePacket(3)
}

func (p *PacketPeerListRequest) SetID() {
	p.BasePacket = NewBasePacket(20)
}

func (p *PacketPeerListSync) SetID() {
	p.BasePacket = NewBasePacket(21)
}

func (p *PacketNewCrate) SetID() {
	p.BasePacket = NewBasePacket(100)
}

func (p *PacketSyncCrates) SetID() {
	p.BasePacket = NewBasePacket(101)
}

func (p *PacketSearchCrates) SetID() {
	p.BasePacket = NewBasePacket(102)
}
