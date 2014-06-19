package cliod

import "log"
import "strings"

type PeerTable struct {
	Table map[[20]byte]Peer
}

func NewPeerTable() PeerTable {
	return PeerTable{make(map[[20]byte]Peer)}
}

type PeerRoute []Peer

func (pr *PeerRoute) String() string {
	var yolo []string
	for _, obj := range *pr {
		yolo = append(yolo, string(obj.ID[:]))
	}
	return strings.Join(yolo, ", ")
}

type Peer struct {
	ID    [20]byte
	Peers []Peer
}

func (pr *PeerRoute) Contains(i Peer) bool {
	for _, peer := range *pr {
		if peer.ID == i.ID {
			return true
		}
	}
	return false
}

func (pt *PeerTable) BuildRoute(start, end [20]byte) *PeerRoute {

	return nil
}

func (pt *PeerTable) recurseRoute(us, last Peer) {
	routes := make([]PeerRoute, 0)

	var recurse func(last PeerRoute)

	recurse = func(last PeerRoute) {
		log.Printf("Last: %v", string(last[len(last)-1].ID[:]))
		for _, peer := range last[len(last)-1].Peers {
			log.Printf("peer: %v", string(peer.ID[:]))
			if last.Contains(peer) {
				continue
			}

			var x PeerRoute
			x = append(last, peer)
			routes = append(routes, x)
			log.Printf("RES: %v", x.String())
			recurse(x)
		}
	}

	recurse(PeerRoute{us})

	log.Printf("# of routes: %v", len(routes))

	for _, route := range routes {
		var yolo []string
		for _, obj := range route {
			yolo = append(yolo, string(obj.ID[:]))
		}
		log.Printf("Routes: %v", strings.Join(yolo, ", "))
	}
}

func New20Byte(last byte) [20]byte {
	var res [20]byte
	for i := 0; i < 19; i++ {
		res[i] = 'x'
	}
	res[19] = byte(last)
	return res
}

func Test() {
	a := Peer{ID: New20Byte('A')}
	b := Peer{ID: New20Byte('B')}
	c := Peer{ID: New20Byte('C')}
	d := Peer{ID: New20Byte('D')}
	e := Peer{ID: New20Byte('E')}
	f := Peer{ID: New20Byte('F')}
	g := Peer{ID: New20Byte('G')}
	h := Peer{ID: New20Byte('H')}
	i := Peer{ID: New20Byte('I')}

	a.Peers = []Peer{b, c}
	b.Peers = []Peer{i}
	c.Peers = []Peer{d}
	d.Peers = []Peer{i}
	e.Peers = []Peer{a, b, c, d, e, f, g, h, i}

	x := NewPeerTable()
	x.Table = map[[20]byte]Peer{
		a.ID: a,
		b.ID: b,
		c.ID: c,
		d.ID: d,
		f.ID: f,
		g.ID: g,
		h.ID: h,
		i.ID: i,
	}

	x.recurseRoute(a, i)
}
