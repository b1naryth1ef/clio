package cliod

import "time"

type Packet interface {
	SetID()
}

type BasePacket struct {
	ID   uint16
	UID  int32
	Time time.Time
	Prop bool
}

type PacketPeer struct {
	ID [20]byte
	IP string
}

func NewBasePacket(ID uint16) BasePacket {
	return BasePacket{
		ID:   ID,
		Time: time.Now(),
		UID:  GetRandomToken(),
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
	PeerCount   int
}

type PacketAuth struct {
	BasePacket
	PublicKey []byte
	T1        int32
	PeerCount int
}

type PacketPing struct {
	BasePacket
	Token int
}

type PacketProxyPacket struct {
	BasePacket
	Dest      [20]byte
	Payload   []byte
	Queue     bool
	Important bool
}

type PacketFindPeer struct {
	BasePacket
	Peer [20]byte
}

type PacketFoundPeer struct {
	BasePacket
	Peer [20]byte
	IP   string
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

func (p *PacketProxyPacket) SetID() {
	p.BasePacket = NewBasePacket(4)
}

func (p *PacketFindPeer) SetID() {
	p.BasePacket = NewBasePacket(10)
}

func (p *PacketFoundPeer) SetID() {
	p.BasePacket = NewBasePacket(11)
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
