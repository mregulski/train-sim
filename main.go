package main

import (
	network "github.com/mregulski/ppt-6-concurrent/network"
	"log"
	"sync"
)

func main() {
	var graph *network.Graph
	log.SetFlags(log.LstdFlags|log.Lmicroseconds)
	if net, err := network.LoadGraph("network.json"); err != nil {
		log.Fatal(err)
	} else {
		graph = net
	}
	log.Printf("%+v\n", graph.Config)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go graph.InfoHandler()

	for _, junction := range graph.Junctions {
		wg.Add(1)
		go network.Handle(junction, graph)
	}

    visited := make(map[network.TrackKeyType]bool)
	for k, tracks := range graph.Tracks {
        if (visited[network.TrackKeyType{k.A,k.B}] || visited[network.TrackKeyType{k.B,k.A}]) {
            continue
        }
		for _, track := range tracks {
			wg.Add(1)
			go network.Handle(track, graph)
		}
        visited[k] = true
	}


	for i := 0; i < len(graph.Vehicles); i++ {
		wg.Add(1)
		go graph.Vehicles[i].Handle(graph)
	}

	// go graph.Vehicles[5].Handle(graph)
	wg.Wait()
}

