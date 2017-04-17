package main

import (
	"fmt"
)

func main() {
	
	net := Network{}
	net.LoadFromJSONFile("network.json")
	fmt.Println(net)
}
