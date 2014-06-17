package cliod

import "os"
import "code.google.com/p/go.crypto/openpgp"

// A ring is a set of public/private keys
type Ring struct {
	Public, Private openpgp.EntityList
}

type ScanComperator func(i *openpgp.Identity) bool

func (r *Ring) scanByEmail(comp ScanComperator, li openpgp.EntityList) *openpgp.Entity {
	for _, ent := range li {
		for _, ident := range ent.Identities {
			if comp(ident) {
				return ent
			}
		}
	}
	return nil
}

func (r *Ring) PubkeyByEmail(email string) *openpgp.Entity {
	return r.scanByEmail(func(i *openpgp.Identity) bool {
		if i.UserId.Email == email {
			return true
		}
		return false
	}, r.Public)
}

func (r *Ring) PrivkeyByEmail(email string) *openpgp.Entity {
	return r.scanByEmail(func(i *openpgp.Identity) bool {
		if i.UserId.Email == email {
			return true
		}
		return false
	}, r.Private)
}

func OpenRing(pub, priv string) Ring {
	pubF, _ := os.Open(pub)
	privF, _ := os.Open(priv)

	pubR, _ := openpgp.ReadKeyRing(pubF)
	privR, _ := openpgp.ReadKeyRing(privF)

	return Ring{
		Public:  pubR,
		Private: privR,
	}
}
