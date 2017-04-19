package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	network "github.com/mregulski/ppt-6-concurrent/network"
)

func main() {
	net := network.Network{}
	if err := net.LoadFromJSONFile("network.json"); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(net)

	wg := &sync.WaitGroup{}
	userInput := make(chan string)
	wg.Add(1)
	go supervisor(net, userInput, wg)
	go func() {
		// var input string
		in := bufio.NewReader(os.Stdin)
		for {
			line, _ := in.ReadString('\n')
			log.Print("User:", line)
			userInput <- line
		}
	}()
	wg.Wait()
}

func supervisor(graph network.Network, userInput <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	queue := make(chan network.Event)
	vehicleControllers := make([]chan interface{}, len(graph.Vehicles))

	for i, v := range graph.Vehicles {
		wg.Add(1)
		vehicleControllers[i] = make(chan interface{})
		station := graph.GetStation(v.Route[0])
		v.UpdatePosition(station.GetFreeTrack().(network.Place))
		go v.Start(vehicleControllers[i], queue, wg)
	}
	for {
		select {
		case msg := <-queue:
			log.Println(msg)
		case cmd := <-userInput:
			for _, recv := range vehicleControllers {
				recv <- cmd
			}
		}
	}
}
