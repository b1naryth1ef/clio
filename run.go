package main

//import "./clicli"
import "./cliod"
import "fmt"
import "time"
import "os"
import "strconv"

var network_id = string("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")

func main() {
	ring := cliod.OpenRing("/home/andrei/.gnupg/pubring.gpg", "/home/andrei/.gnupg/secring.gpg")
	fmt.Printf("%s", ring)
	key := ring.PrivkeyByEmail("b1naryth1ef@gmail.com")
	fmt.Printf("%s\n", key)

	// store := cliod.NewStore("/home/andrei/.clio")
	// store.Init()

	// crate := cliod.NewCrate([]byte{}, []string{"test"})
	// store.PutCrate(crate)

	//fmt.Printf("Results: %v\n", len(store.Index.FindByTags([]string{"test"})))

	our_port, _ := strconv.Atoi(os.Args[1])
	their_port, _ := strconv.Atoi(os.Args[2])

	client := cliod.NewNetClient(our_port, key, &ring)
	go client.ServerListenerLoop()

	client.Seed(network_id, []string{
		fmt.Sprintf("127.0.0.1:%v", their_port),
	})

	for {
		time.Sleep(time.Second * 5)
	}

	//cli := cliod.ClientLogin(key, "")
	//cliod.BuildPacketHello(key.PrimaryKey.Fingerprint, [32]byte{""})
}
