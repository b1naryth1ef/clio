package cliod

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func GetCurrentUserHome() string {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd := exec.Command("whoami")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Run()
		cmd.Wait()
		return "/home/" + strings.Replace(out.String(), "\n", "", -1) + "/"
	}
	return ""
}

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

func NewRandomID() string {
	return strings.Replace(uuid.New(), "-", "", -1)
}
