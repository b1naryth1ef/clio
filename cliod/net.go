package cliod

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	//"strings"
)

type PacketID uint16

type NetNode struct {
	ID    [20]byte
	Trust int
	Conn  net.Conn
	Valid bool

	tempToken int32
}

type NetClient struct {
	Ident     *openpgp.Entity
	Ring      openpgp.KeyRing
	NetworkID string
	Outgoing  map[[20]byte]*NetNode
	Incoming  map[[20]byte]*NetNode
	IncQ      []*NetNode
	Server    net.Listener
}

// Create a new NetClient
func NewNetClient(listen_port int, id *openpgp.Entity, r openpgp.KeyRing) NetClient {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", listen_port))
	if err != nil {
		log.Panicf("Error binding: %v", err)
	}

	return NetClient{
		Ident:    id,
		Ring:     r,
		Outgoing: make(map[[20]byte]*NetNode),
		Incoming: make(map[[20]byte]*NetNode),
		IncQ:     make([]*NetNode, 0),
		Server:   server,
	}
}

func (nn *NetNode) ListenLoop(nc *NetClient) {
	var id_pk BasePacket

	reader := bufio.NewReader(nn.Conn)
	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			log.Printf("Error: %v", err)
			break
		}
		data[len(data)-1] = ' '

		json.Unmarshal(data, &id_pk)

		log.Printf("Got packet w/ id `%v`", id_pk.ID)
		switch id_pk.ID {
		case 1:
			var obj PacketHello
			json.Unmarshal(data, &obj)
			nn.handleWelcomePacket(nc, obj)
		case 2:
			var obj PacketAuth
			json.Unmarshal(data, &obj)
			nn.handleAuthPacket(nc, obj)
		}
	}
}

func (nn *NetNode) handleAuthPacket(nc *NetClient, p PacketAuth) {
	log.Printf("uw0tm8?")
	r := bytes.NewReader(p.Payload)
	dec, e := openpgp.ReadMessage(r, nc.Ring, func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		priv := keys[0].PrivateKey
		if priv.Encrypted {
			log.Printf("FUCK")
			priv.Decrypt([]byte(""))
		}
		buf := new(bytes.Buffer)
		priv.Serialize(buf)
		return buf.Bytes(), nil
	}, nil)

	log.Printf("%v", dec.IsEncrypted)
}

func (nn *NetNode) handleWelcomePacket(nc *NetClient, p PacketHello) {
	if p.NetworkHash != nc.NetworkID {
		log.Printf("Network ID of welcome packet does not match (`%v` vs `%v`)",
			p.NetworkHash, nc.NetworkID)
		return
	}

	// Load Public Key
	pubr := bytes.NewReader([]byte(p.PublicKey))
	pub_ring, err := openpgp.ReadKeyRing(pubr)

	if len(pub_ring) < 1 {
		log.Printf("ERROR HANDLING WELCOME PACKET: No public keys decoded!")
		return
	}

	packet := PacketAuth{}

	data, _ := json.Marshal(PacketAuthPayload{p.Token, 1})
	rdata := bytes.NewReader(data)

	// TODO: sign that motherfucker
	var edata bytes.Buffer
	ptxt, err := openpgp.Encrypt(&edata, pub_ring, nil, nil, nil)
	if err != nil {
		log.Printf("ERROR ENCRYPTING: %v", err)
	}

	io.Copy(ptxt, rdata)
	ptxt.Close()

	pub_w := bytes.NewBuffer([]byte{})
	nc.Ident.Serialize(pub_w)
	packet.ID = 2
	packet.PublicKey = pub_w.Bytes()
	packet.Payload = edata.Bytes()

	nn.Send(&packet)
}

// Send a packet to the node
func (nn *NetNode) Send(packet Packet) {
	data, _ := json.Marshal(packet)
	data = append(data, '\n')
	nn.Conn.Write(data)
}

func (nn *NetNode) Handshake(nc *NetClient) bool {
	// Write the public key too a buffer
	pkbuff := bytes.NewBuffer(make([]byte, 0))
	nc.Ident.Serialize(pkbuff)

	nn.tempToken = GetRandomToken()
	nn.Send(&PacketHello{
		ID:          1,
		PublicKey:   pkbuff.Bytes(),
		NetworkHash: nc.NetworkID,
		Token:       nn.tempToken,
	})
	return true
}

func (nc *NetClient) ServerListenerLoop() {
	for {
		conn, err := nc.Server.Accept()
		if err != nil {
			log.Panicf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("Got connection: %v", conn.RemoteAddr())
		node := NetNode{
			Conn: conn,
		}

		nc.IncQ = append(nc.IncQ, &node)
		go node.ListenLoop(nc)
	}
}

func (nc *NetClient) Seed(netid string, ips []string) {
	nc.NetworkID = netid
	for _, ip := range ips {
		nn := nc.AttemptConnection(ip)
		if nn == nil {
			continue
		}

		if nn.Valid {
			nc.Outgoing[nn.ID] = nn
		}
	}
}

func (nc *NetClient) AttemptConnection(ip string) *NetNode {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Printf("Error dialing: %v", err)
		return nil
	}

	node := NetNode{
		ID:    [20]byte{},
		Trust: 0,
		Conn:  conn,
	}

	node.Handshake(nc)
	go node.ListenLoop(nc)

	return &node
}
