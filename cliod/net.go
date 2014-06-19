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

var MAX_HISTORY int = 50

// A NetNode represents an external network node
type NetNode struct {
	// ID represents the 20 byte public key fingerprint
	ID [20]byte

	// The level of trust we have assigned this node (unused currently)
	Trust int

	// The actual socket too this client
	Conn net.Conn

	// Whether the client is authenticated yet
	Authed bool

	// The amount of peers this client has
	PeerCount int

	// Represents the last time this node asked us to propagate something for limiting
	LastProp time.Time

	// The public key of thise node
	Key openpgp.EntityList

	nc    *NetClient
	token int32
}

// A NetClient represents a local client
type NetClient struct {
	// Our local public/private key pair
	Ident *openpgp.Entity

	// The ring for decrypting messages (TODO: try creating a new ring w/ ident)
	Ring *Ring

	// The network ID
	NetworkID string

	// A mapping of peer fingerprint/ids too the actual nodes
	Peers map[[20]byte]*NetNode

	// A socket server
	Server net.Listener

	// The backend crate-store
	Store *Store

	// List of historical packets we've parsed
	history []int32
}

// Create a new NetClient
func NewNetClient(listen_port int, id *openpgp.Entity, r *Ring, s *Store) NetClient {
	server, err := net.Listen("tcp", fmt.Sprintf(":%d", listen_port))
	if err != nil {
		log.Panicf("Error binding: %v", err)
	}

	return NetClient{
		Ident:  id,
		Ring:   r,
		Peers:  make(map[[20]byte]*NetNode),
		Server: server,
		Store:  s,
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
	reader := bufio.NewReader(nn.Conn)
	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			log.Printf("Error: %v", err)
			delete(nc.Peers, nn.ID)
			break
		}
		data[len(data)-1] = ' '

		nn.HandlePacket(nc, data)

	}
}

func (nn *NetNode) HandlePacket(nc *NetClient, data []byte) {
	var id_pk BasePacket
	var parsed Packet

	// If the data starts with a null byte, it's encrypted
	if data[0] == '\x00' {
		var obj EncryptedPacket

		// Unmarshal the data into the EncryptedPacket struct, ignoring the null byte
		json.Unmarshal(data[1:], &obj)

		// Decrypt the payload (this may fail if it's not encrypted for us)
		dec := nc.Decrypt(obj.Payload)
		if dec == nil {
			log.Printf("ERROR: Failed to decrypt packet!")
			return
		}

		// Make sure the payload was actually encrypted and signed by someone, this prevents people
		//  from just sending random shit in the payload
		if !dec.IsEncrypted || !dec.IsSigned {
			log.Printf("ERROR: Encrypted Packet was not signed, or not encrypted!")
			return
		}

		// Make sure that this packet was encrypted from the node that it was sent from
		// TOOD: in the future we may want to allow proxying encrypted messages
		sfp := dec.SignedBy.PublicKey.Fingerprint
		if nn.ID != [20]byte{} && sfp != nn.ID {
			log.Printf("ERROR: Encrypted Packet was signed by a different key then the nodes ID! (%v != %v)",
				dec.SignedBy.PublicKey.Fingerprint, nn.ID)
			return
		}

		// Make sure we have data, and replace the data buffer with the decrypted data
		data, _ = ioutil.ReadAll(dec.LiteralData.Body)
		if len(data) <= 0 {
			log.Printf("ERROR: No data decoded from encrypted packet!")
			return
		}
	}

	// Unmarshal the data into a blank packet struct to get the type out
	json.Unmarshal(data, &id_pk)

	// Check if the timestamp has expired (this will be encrypted, and thus prevents replay attacks)
	if time.Now().Sub(id_pk.Time) > (time.Second * 30) {
		log.Printf("Packet is expired: %v (%v)", id_pk.Time, data)
		return
	}

	log.Printf("Got packet w/ id `%v`", id_pk.ID)

	// Make sure we haven't seen this /exact/ packet before
	if nc.InHistory(id_pk.UID) {
		log.Printf("Already parsed packet w/ ID %v")
		return
	}

	// Parse the packet into a struct and pass it too a handler function
	switch id_pk.ID {
	case 1:
		var obj PacketHello
		json.Unmarshal(data, &obj)
		nn.handleWelcomePacket(nc, obj)
		parsed = &obj
	case 2:
		var obj PacketAuth
		json.Unmarshal(data, &obj)
		nn.handleAuthPacket(nc, obj)
		parsed = &obj
	case 3:
		var obj PacketPing
		json.Unmarshal(data, &obj)
		nn.handlePingPacket(nc, obj)
		parsed = &obj
	case 20:
		var obj PacketPeerListRequest
		json.Unmarshal(data, &obj)
		nn.handlePeerListRequestPacket(nc, obj)
		parsed = &obj
	case 21:
		var obj PacketPeerListSync
		json.Unmarshal(data, &obj)
		nn.handlePeerListSyncPacket(nc, obj)
		parsed = &obj
	}

	// Add packet too history, and slice end of history off if required
	nc.history = append(nc.history, id_pk.UID)
	if len(nc.history) > MAX_HISTORY {
		nc.history = nc.history[1:]
	}

	// If this packet is asking too be propegated, try doing so
	if id_pk.Prop {
		// Rate limiting for propagation
		if time.Now().Sub(nn.LastProp) <= time.Minute {
			log.Printf("Cannot propagate packet w/ id %v from %v; too soon sense last propagate!", id_pk.ID, nn.ID)
			return
		}
		nn.LastProp = time.Now()

		// Send to all the peers we have
		for _, node := range nc.Peers {
			if node == nn {
				continue
			}

			node.Send(parsed, true)
		}
	}
}

func (nn *NetNode) handleProxyPacket(nc *NetClient, p PacketProxyPacket) {
	// Check if the packet is intended for us
	if p.Dest == nc.Ident.PrimaryKey.Fingerprint {
		nn.HandlePacket(nc, p.Payload)
		return
	}

	// Check if we have the destination node as a peer
	if _, exists := nc.Peers[p.Dest]; exists {
		nc.Peers[p.Dest].Send(&p, false)
		return
	}

	// TODO: handle queue
	// TODO: handle importance
}

func (nn *NetNode) handlePingPacket(nc *NetClient, p PacketPing) {
	log.Printf("I'VE BEEN PINGED MATEY!")
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

		// Don't send clients themselves
		if peer == nn {
			continue
		}

		pr := PacketPeer{peer.ID, peer.Conn.RemoteAddr().String()}

		peerli = append(peerli, pr)
	}

	nn.Send(&PacketPeerListSync{
		Peers: peerli,
	}, false)
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
	nn.PeerCount = p.PeerCount

	log.Printf("Successfully authenticated!")
	nn.Send(&PacketPeerListRequest{
		MaxPeers: 50,
	}, false)

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
	nn.PeerCount = p.PeerCount

	// This doesn't quiet seem right
	nc.Peers[nn.ID] = nn

	pub_w := bytes.NewBuffer([]byte{})
	nc.Ident.Serialize(pub_w)

	packet := PacketAuth{
		PublicKey: pub_w.Bytes(),
		T1:        p.Token,
		PeerCount: len(nc.Peers),
	}

	nn.Send(&packet, false)
}

func (nn *NetNode) handleNewCrate(nc *NetClient, p PacketNewCrate) {
	if nc.Store.HasCrate(p.Crate.ID) {
		log.Printf("Not adding crate %v, already have it!", p.Crate.ID)
		return
	}

	log.Printf("Adding crate %v", p.Crate.ID)
	nc.Store.PutCrate(p.Crate)
}

// Build a packet data
func (nn *NetNode) BuildPacket(packet Packet, writeid bool) []byte {
	var network []byte

	if writeid {
		packet.SetID()
	}

	data, err := json.Marshal(packet)
	if err != nil {
		log.Printf("ERROR BUILDING: %v", err)
		return network
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

	return append(network, '\n')
}

// Send a packet to the node
func (nn *NetNode) Send(packet Packet, pass bool) {
	nn.Conn.Write(nn.BuildPacket(packet, !pass))
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
		PeerCount:   len(nc.Peers),
	}, false)
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

func (nc *NetClient) SendToClient(id [20]byte, p Packet, queued bool, important bool) bool {
	if _, exists := nc.Peers[id]; exists {
		nc.Peers[id].Send(p, true)
		return false
	}

	var sendli []*NetNode
	if important {
		for _, node := range nc.Peers {
			sendli = append(sendli, node)
		}
	} else {
		var best *NetNode
		for _, node := range nc.Peers {
			if best == nil || node.PeerCount > best.PeerCount {
				best = node
			}
		}
		sendli = []*NetNode{best}
	}

	packet_data, _ := json.Marshal(p)
	packet := PacketProxyPacket{
		Dest:      id,
		Payload:   packet_data,
		Queue:     queued,
		Important: important,
	}

	for _, node := range sendli {
		node.Send(&packet, true)
	}
	return true
}

func (nc *NetClient) InHistory(id int32) bool {
	for _, val := range nc.history {
		if val == id {
			return true
		}
	}
	return false
}
