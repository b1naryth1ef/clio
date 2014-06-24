package main

//import "./clicli"
import (
	"./cliod"
	"fmt"
	"os"
	"strconv"
	"time"
)

var network_id = string("")

func main() {
	// cliod.Test()
	// return
	if len(os.Args) < 3 {
		fmt.Printf("Usage: ./run <email> <port> [seed_addr]\n")
		return
	}

	user := cliod.GetCurrentUserHome()

	var email, seed string
	var port int

	email = os.Args[1]
	port, _ = strconv.Atoi(os.Args[2])
	if len(os.Args) >= 4 {
		seed = os.Args[3]
	}

	ring := cliod.OpenRing(user+"/.gnupg/pubring.gpg", user+"/.gnupg/secring.gpg")
	key := ring.PrivkeyByEmail(email)

	store := cliod.NewStore(user + "/.clio")
	store.Init()

	client := cliod.NewNetClient(port, key, &ring, &store)
	client.Run()

	if seed != "" {
		client.Seed(network_id, []string{seed})
	}

	for {
		time.Sleep(time.Second * 5)
	}
}
