package main

import "github.com/secondbit/wendy"
import "os"

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err.Error())
	}
	id, err := wendy.NodeIDFromBytes([]byte(hostname))
	if err != nil {
		panic(err.Error())
	}
	node := wendy.NewNode(id, "127.0.0.1", "your_global_ip_address", "your_region", 8080)
}
