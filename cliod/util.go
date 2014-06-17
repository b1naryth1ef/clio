package cliod

import (
	"math/rand"
	"os"
	"time"
)

func GetRandomToken() int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int31()
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
