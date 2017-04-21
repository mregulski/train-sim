package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Vehicle struct {
	ID            int
	MaxSpeed      float64
	Capacity      int // not used yet
	Route         []*Node
	position      Location      // current Place occupied by this Vehicle
	lastPosition  Location      // last Place visited
	routePosition int           // last visited node on the route
	ctrl          <-chan string // channel for receiving commands
	log           *log.Logger
}

func (v Vehicle) String() string {
	var sRoute string
	for i, node := range v.Route {
		sRoute += fmt.Sprintf("%d", node.ID)
		if i < len(v.Route)-1 {
			sRoute += fmt.Sprintf("->")
		}
	}
	return fmt.Sprintf("{ID: %d, MaxSpeed: %f, Capacity: %d, route: %s}",
		v.ID, v.MaxSpeed, v.Capacity, sRoute)
}

func (v *Vehicle) LocationName() string {
	return v.position.Name()
}

func (v *Vehicle) UpdatePosition(newLocation Location) {
	for {
		if newLocation.Take(v) {
			if v.position != nil {
				v.position.Leave()
			}
			v.lastPosition = v.position
			v.position = newLocation
			return
		}
		wait := ScaledTime(rand.Intn(10)) //)time.Millisecond * time.Duration(rand.Intn(10)) * time.Duration(GetConfig().TimeScale)
		v.log.Printf("Destination %s not available, sleeping %s", newLocation.Name(), wait)
		time.Sleep(wait)
	}
	v.log.Printf("new position: %s\n", newLocation.Name())
}

//NextLocation returns the vehicle's new position
func (v *Vehicle) NextLocation() Location {
	var destination Location
	switch pos := v.position.(type) {
	default:
		v.log.Fatalf("Unexpected type %T\n", pos)
		destination = pos
	case Track:
		// destination is the other end of the track
		v.routePosition = (v.routePosition + 1) % len(v.Route)
		ends := pos.EndPoints()
		if v.lastPosition == ends[0] {
			destination = ends[1]
		} else {
			destination = ends[0]
		}
	case *Node:
		// destination is the first free track from current node to the next one on route
		target := v.Route[(v.routePosition+1)%len(v.Route)]
		tracks := TrackCollection(pos.Tracks[trackKey(target.ID)])
		if len(tracks) == 0 {
			v.log.Fatalf("No tracks found between <%d> and <%d>\n", pos.ID, target.ID)
		}
		for {
			destination = tracks.GetFreeTrack()
			if destination != nil {
				break
			}
			v.log.Printf("No free track between %s and %s, waiting...\n", pos.Name(), target.Name())
		}
	}

	return destination
}

func ScaledTime(baseTime int) time.Duration {
	return time.Millisecond * time.Duration(baseTime*GetConfig().TimeScale)
}

/*
DoRound - simulate vehicle's behaviour for a round
*/
func (v *Vehicle) DoRound() error {
	var eta = ScaledTime(v.position.TravelTime(v.MaxSpeed))
	time.Sleep(eta)

	destination := v.NextLocation()
	v.UpdatePosition(destination)

	v.log.Println(v.lastPosition.Name(), "->", v.position.Name())

	return nil
}

/*
Start - begin simulation of this vehicle
*/
func (v *Vehicle) Start(start Location, controller <-chan string, queue chan Event, wg *sync.WaitGroup) {
	v.position = start
	v.log.Println("[info] Starting")
	v.log.Printf("[info] Initial position: %s\n", v.position)
	defer wg.Done()
	for {
		select {
		case cmd := <-controller:
			log.Printf("[info] Received: '%s'\n", cmd)
			if cmd == "quit" {
				log.Printf("[cmd] %s - stopping the vehicle\n", cmd)
				return
			}
		default:
			v.DoRound()
		}
	}
}

func vehicleFromJSON(rawVehicle map[string]*json.RawMessage, nodes []*Node) *Vehicle {
	var routeIDs []int
	var vehicle Vehicle
	json.Unmarshal(*rawVehicle["id"], &vehicle.ID)
	json.Unmarshal(*rawVehicle["maxSpeed"], &vehicle.MaxSpeed)
	json.Unmarshal(*rawVehicle["capacity"], &vehicle.Capacity)
	json.Unmarshal(*rawVehicle["route"], &routeIDs)
	for _, nodeID := range routeIDs {
		vehicle.Route = append(vehicle.Route, nodes[nodeID-1])
	}
	vehicle.lastPosition = vehicle.Route[0]
	vehicle.log = log.New(os.Stdout, fmt.Sprintf("[vehicle:%d] ", vehicle.ID), log.Ltime|log.Lmicroseconds)
	if !logging {
		vehicle.log.SetOutput(ioutil.Discard)
	}
	return &vehicle
}

type NopLogger struct {
	*log.Logger
}

func (l *NopLogger) Print(v ...interface{})                 { /*nop*/ }
func (l *NopLogger) Printf(format string, v ...interface{}) { /*nop*/ }
func (l *NopLogger) Println(v ...interface{})               { /*nop*/ }
