package cliod

import (
	"math/rand"
	"time"
)

func GetRandomToken() int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int31()
}
