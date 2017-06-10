package main

import (
	network "github.com/mregulski/ppt-6-concurrent/network"
	"log"
	"fmt"
	"time"
)

func main() {
	var graph *network.Graph
	log.SetFlags(log.LstdFlags|log.Lmicroseconds)
	if net, err := network.LoadGraph("network.json"); err != nil {
		log.Fatal(err)
	} else {
		graph = net
	}
	fmt.Printf("%+v\n", graph.Config)

	fmt.Printf("\n----------\nJunctions\n----------\n")
	for _, junction := range graph.Junctions {
		fmt.Printf("%v, station: %s\n", junction, graph.StationWith(junction))
	}

	fmt.Printf("\n----------\nStations\n----------\n")
	for _, station := range graph.Stations {
		fmt.Printf("%v\n", station)
	}

	fmt.Printf("\n----------\nTracks\n----------\n")
	for _, track := range graph.Tracks() {
			fmt.Printf("%v\n", track)
	}

	fmt.Printf("\n----------\nVehicles\n----------\n")
	for _, vehicle := range graph.Vehicles {
			fmt.Printf("%v\n", vehicle)
	}

	fmt.Printf("\n\n====================================\nStarting simulation\n====================================\n\n")
	<-time.After(time.Second)
	graph.Start()
}

