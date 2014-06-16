package main

//import "./clicli"
import "./cliod"
import "fmt"

func main() {
	ring := cliod.OpenRing("/home/b1nzy/.gnupg/pubring.gpg", "/home/b1nzy/.gnupg/secring.gpg")
	fmt.Printf("%s", ring)
	fmt.Printf("%s\n", ring.FindByEmail("b1naryth1ef@gmail.com"))
}
