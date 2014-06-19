package main

//import "./clicli"
import (
	"./cliod"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"
)

var network_id = string("")

func main() {
	usr, _ := user.Current()
	ring := cliod.OpenRing(usr.HomeDir+"/.gnupg/pubring.gpg", usr.HomeDir+"/.gnupg/secring.gpg")
	key := ring.PrivkeyByEmail("b1naryth1ef+1@gmail.com")

	store := cliod.NewStore(usr.HomeDir + "/.clio")
	store.Init()

	// crate := cliod.NewCrate([]byte{}, []string{"test"})
	// store.PutCrate(crate)

	// fmt.Printf("Results: %v\n", len(store.Index.FindByTags([]string{"test"})))

	our_port, _ := strconv.Atoi(os.Args[1])

	client := cliod.NewNetClient(our_port, key, &ring, &store)
	go client.ServerListenerLoop()

	if len(os.Args) > 2 {
		their_port, _ := strconv.Atoi(os.Args[2])
		client.Seed(network_id, []string{
			fmt.Sprintf("127.0.0.1:%v", their_port),
		})
	}

	for {
		time.Sleep(time.Second * 5)
	}

	//cli := cliod.ClientLogin(key, "")
	//cliod.BuildPacketHello(key.PrimaryKey.Fingerprint, [32]byte{""})
}
