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
	"time"
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
	Peers     map[[20]byte]*NetNode
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
		Ident:  id,
		Ring:   r,
		Peers:  make(map[[20]byte]*NetNode),
		Q:      make([]*NetNode, 0),
		Server: server,
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
		log.Printf("ERROR: %v", e)
	}

	return dec
}

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

			dec := nc.Decrypt(obj.Payload)
			if dec == nil {
				log.Printf("ERROR: Failed to decrypt packet!")
				continue
			}

			if !dec.IsEncrypted || !dec.IsSigned {
				log.Printf("ERROR: Encrypted Packet was not signed, or not encrypted!")
				continue
			}

			log.Printf("Signed By: %v", dec.SignedBy.PublicKey.Fingerprint)
			sfp := dec.SignedBy.PublicKey.Fingerprint
			if nn.ID != [20]byte{} && sfp != nn.ID {
				log.Printf("ERROR: Encrypted Packet was signed by a different key then the nodes ID! (%v != %v)",
					dec.SignedBy.PublicKey.Fingerprint, nn.ID)
				continue
			}

			data, _ = ioutil.ReadAll(dec.LiteralData.Body)
			if len(data) <= 0 {
				log.Printf("ERROR: No data decoded from encrypted packet!")
				return
			}
		}

		json.Unmarshal(data, &id_pk)

		if time.Now().Sub(id_pk.Time) > (time.Second * 30) {
			log.Printf("Packet is expired: %v (%v)", id_pk.Time, data)
			continue
		}

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
		case 3:
			var obj PacketPing
			json.Unmarshal(data, &obj)
			nn.handlePingPacket(nc, obj)
		case 20:
			var obj PacketPeerListRequest
			json.Unmarshal(data, &obj)
			nn.handlePeerListRequestPacket(nc, obj)
		case 21:
			var obj PacketPeerListSync
			json.Unmarshal(data, &obj)
			nn.handlePeerListSyncPacket(nc, obj)
		}

	}
}

func (nn *NetNode) handlePingPacket(nc *NetClient, p PacketPing) {

}

func (nn *NetNode) handlePeerListRequestPacket(nc *NetClient, p PacketPeerListRequest) {
	// Prepare a peer list and send it
	var i int32
	peerli := make([]PacketPeer, 0)

	for _, peer := range nc.Peers {
		i += 1
		if i > p.MaxPeers {
			break
		}

		pr := PacketPeer{peer.ID, peer.Conn.RemoteAddr().String()}

		peerli = append(peerli, pr)
	}

	nn.Send(&PacketPeerListSync{
		Peers: peerli,
	})
}

func (nn *NetNode) handlePeerListSyncPacket(nc *NetClient, p PacketPeerListSync) {
	for _, peer := range p.Peers {
		if _, exists := nc.Peers[peer.ID]; !exists {
			log.Printf("Would try to connect to %v (%v)", peer.IP, peer.ID)
		} else {
			log.Printf("Won't connect too %v (%v), already have them!", peer.IP, peer.ID)
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

	pubr := bytes.NewReader(p.PublicKey)
	nn.Key, _ = openpgp.ReadKeyRing(pubr)

	// ehhhhhh
	nn.ID = nn.Key[0].PrimaryKey.Fingerprint
	nc.Peers[nn.ID] = nn

	log.Printf("Successfully authenticated!")
	nn.Send(&PacketPeerListRequest{
		MaxPeers: 50,
	})

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
	nn.ID = nn.Key[0].PrimaryKey.Fingerprint

	// This doesn't quiet seem right
	nc.Peers[nn.ID] = nn

	pub_w := bytes.NewBuffer([]byte{})
	nc.Ident.Serialize(pub_w)

	packet := PacketAuth{
		PublicKey: pub_w.Bytes(),
		T1:        p.Token,
	}

	nn.Send(&packet)
}

// Send a packet to the node
func (nn *NetNode) Send(packet Packet) {
	var network []byte

	packet.SetID()
	data, err := json.Marshal(packet)
	if err != nil {
		log.Printf("ERROR SENDING: %v", err)
		return
	}

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

		go node.ListenLoop(nc)
	}
}

func (nc *NetClient) Seed(netid string, ips []string) {
	nc.NetworkID = netid
	for _, ip := range ips {
		nc.AttemptConnection(ip)
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
		nc:    nc,
	}

	go node.ListenLoop(nc)
	node.Handshake(nc)

	return &node
}
