package cliod

import (
	"bytes"
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
