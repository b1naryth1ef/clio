package main

//import "./clicli"
import "./cliod"
import "fmt"
import "time"
import "os"
import "strconv"

var network_id = string("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")

func main() {
	ring := cliod.OpenRing("/home/b1nzy/.gnupg/pubring.gpg", "/home/b1nzy/.gnupg/secring.gpg")
	fmt.Printf("%s", ring)
	key := ring.PubkeyByEmail("b1naryth1ef+1@gmail.com")
	fmt.Printf("%s\n", key)

	i, _ := strconv.Atoi(os.Args[1])
	client := cliod.NewNetClient(i, key, ring.Private)
	go client.ServerListenerLoop()

	client.Seed(network_id, []string{
		"127.0.0.1:1337",
	})

	for {
		time.Sleep(time.Second * 5)
	}

	//cli := cliod.ClientLogin(key, "")
	//cliod.BuildPacketHello(key.PrimaryKey.Fingerprint, [32]byte{""})
}
