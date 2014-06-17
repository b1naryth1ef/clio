package cliod

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

// A NetNode represents an external network node
type NetNode struct {
	ID     [20]byte
	Trust  int
	Conn   net.Conn
	Valid  bool
	Authed bool

	Key openpgp.EntityList

	nc    *NetClient
	token int32
}

// A NetClient represents a local client
type NetClient struct {
	Ident     *openpgp.Entity
	Ring      *Ring
	NetworkID string
	Outgoing  map[[20]byte]*NetNode
	Incoming  map[[20]byte]*NetNode
	Q         []*NetNode
	Server    net.Listener
}

// Create a new NetClient
func NewNetClient(listen_port int, id *openpgp.Entity, r *Ring) NetClient {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", listen_port))
	if err != nil {
		log.Panicf("Error binding: %v", err)
	}

	return NetClient{
		Ident:    id,
		Ring:     r,
		Outgoing: make(map[[20]byte]*NetNode),
		Incoming: make(map[[20]byte]*NetNode),
		Q:        make([]*NetNode, 0),
		Server:   server,
	}
}

func (nc *NetClient) Decrypt(data []byte) *openpgp.MessageDetails {
	r := bytes.NewReader(data)

	dec, e := openpgp.ReadMessage(r, nc.Ring.Private, func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		priv := keys[0].PrivateKey
		if priv.Encrypted {
			priv.Decrypt([]byte(""))
		}
		buf := new(bytes.Buffer)
		priv.Serialize(buf)
		return buf.Bytes(), nil
	}, nil)

	if e != nil {
		log.Printf("%v", e)
	}

	return dec
}

// TODO: replay attack
func (nc *NetClient) Encrypt(data []byte, to openpgp.EntityList) []byte {
	var edata bytes.Buffer
	ptxt, err := openpgp.Encrypt(&edata, to, nc.Ident, nil, nil)
	if err != nil {
		log.Printf("ERROR ENCRYPTING: %v", err)
	}

	rdata := bytes.NewReader(data)

	io.Copy(ptxt, rdata)
	ptxt.Close()

	return edata.Bytes()
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

		// If the data starts with a null byte, it's encrypted
		if data[0] == '\x00' {
			var obj EncryptedPacket

			json.Unmarshal(data[1:], &obj)

			log.Printf("S3: %v", len(obj.Payload))
			dec := nc.Decrypt(obj.Payload)
			if dec == nil {
				log.Printf("ERROR: Failed to decrypt packet!")
				return
			}

			if !dec.IsEncrypted || !dec.IsSigned {
				log.Printf("ERROR: Encrypted Packet was not signed, or not encrypted!")
				return
			}

			if nn.ID != [20]byte{} && dec.SignedBy.PublicKey.Fingerprint != nn.ID {
				log.Printf("ERROR: Auth Packet was signed by a different key then the nodes ID! (%v != %v)",
					dec.SignedBy.PublicKey.Fingerprint, nn.ID)
				return
			}

			data, _ = ioutil.ReadAll(dec.LiteralData.Body)
		}

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
	if p.T1 != nn.token {
		log.Printf("ERROR: Auth Packet token did not match!")
		return
	}

	// We are atuhenticated, and will start using the new token
	nn.Authed = true
	nn.token = p.T2

	pubr := bytes.NewReader(p.PublicKey)
	nn.Key, _ = openpgp.ReadKeyRing(pubr)

	// ehhhhhh
	nn.ID = nn.Key[0].PrimaryKey.Fingerprint

	log.Printf("Successfully authenticated!")
}

func (nn *NetNode) handleWelcomePacket(nc *NetClient, p PacketHello) {
	if p.NetworkHash != nc.NetworkID {
		log.Printf("Network ID of welcome packet does not match (`%v` vs `%v`)",
			p.NetworkHash, nc.NetworkID)
		return
	}

	// Load Public Key
	pubr := bytes.NewReader(p.PublicKey)
	pub_ring, _ := openpgp.ReadKeyRing(pubr)

	if len(pub_ring) < 1 {
		log.Printf("ERROR HANDLING WELCOME PACKET: No public keys decoded!")
		return
	}

	nn.Key = pub_ring

	pub_w := bytes.NewBuffer([]byte{})
	nc.Ident.Serialize(pub_w)

	packet := PacketAuth{
		ID:        2,
		PublicKey: pub_w.Bytes(),
		T1:        p.Token,
		T2:        nn.token,
	}

	nn.Send(&packet)
}

// Send a packet to the node
func (nn *NetNode) Send(packet Packet) {
	var network []byte
	data, _ := json.Marshal(packet)

	if nn.Key != nil {
		network = append(network, '\x00')

		data = nn.nc.Encrypt(data, nn.Key)

		data, _ = json.Marshal(EncryptedPacket{
			Payload: data,
		})

		for _, b := range data {
			network = append(network, b)
		}
	} else {
		network = data
	}

	network = append(network, '\n')

	nn.Conn.Write(network)
}

func (nn *NetNode) Handshake(nc *NetClient) bool {
	// Write the public key too a buffer
	pkbuff := bytes.NewBuffer(make([]byte, 0))
	nc.Ident.Serialize(pkbuff)

	nn.token = GetRandomToken()
	nn.Send(&PacketHello{
		ID:          1,
		PublicKey:   pkbuff.Bytes(),
		NetworkHash: nc.NetworkID,
		Token:       nn.token,
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
			nc:   nc,
			Conn: conn,
		}

		nc.Q = append(nc.Q, &node)
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
