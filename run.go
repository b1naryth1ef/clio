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

	i, _ := strconv.Atoi(os.Args[1])
	client := cliod.NewNetClient(i, key, &ring)
	go client.ServerListenerLoop()

	client.Seed(network_id, []string{
		"127.0.0.1:1338",
	})

	for {
		time.Sleep(time.Second * 5)
	}

	//cli := cliod.ClientLogin(key, "")
	//cliod.BuildPacketHello(key.PrimaryKey.Fingerprint, [32]byte{""})
}
