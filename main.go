package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	network "github.com/mregulski/ppt-6-concurrent/network"
	"strings"
	"flag"
)

var logging = true

func main() {
	fInteractive := flag.Bool("interactive", false, "if present, runs n interactive mode (no logging)")
	flag.Parse()
	logging = !*fInteractive
	net := network.Network{}
	network.SetLogging(logging)
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
			line, err := in.ReadString('\n')
			if err != nil {
				log.Println("[user] error reading user input:", err)
				continue
			}
			line = strings.TrimSpace(line)
			log.Printf("[user] '%s'\n", line)
			userInput <- line
		}
	}()
	wg.Wait()
}

func supervisor(graph network.Network, userInput <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	queue := make(chan network.Event)
	vehicleControllers := make([]chan string, len(graph.Vehicles))
	l := log.New(os.Stdout, "[supervisor] ", 0)
	for i, v := range graph.Vehicles {
		wg.Add(1)
		vehicleControllers[i] = make(chan string)
		station := graph.GetStation(v.Route[0])
		start := station.GetFreeTrack().(network.Location)
		start.Take(v)
		go v.Start(start, vehicleControllers[i], queue, wg)
	}
	for {
		select {
		case msg := <-queue:
			log.Println(msg)
		case cmd := <-userInput:
			if cmd == "list" {
				for _, v := range graph.Vehicles {
					fmt.Printf("{ID: %d, position: %s}\n", v.ID, v.LocationName()	)
				}
			}
			if cmd == "quit" {
				l.Printf("received '%s' - stopping simulation\n", cmd)
				for _, recv := range vehicleControllers {
					recv <- cmd
						
				}
				return
			}
		}
	}
}
