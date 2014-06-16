package cliod

import "os"
import "code.google.com/p/go.crypto/openpgp"

type Ring struct {
	Public, Private openpgp.EntityList
}

func (r *Ring) FindByEmail(email string) *openpgp.Entity {
	for _, ent := range r.Private {
		for _, ident := range ent.Identities {
			if ident.UserId.Email == email {
				return ent
			}
		}
	}
	return nil
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
