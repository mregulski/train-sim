package network

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type Vehicle struct {
	ID       int64
	MaxSpeed float64
	Capacity int64
	Route    []*Node
	position Place
	timeLeft float64
	ctrl     <-chan string
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

func (v *Vehicle) logf(format string, args ...interface{}) {
	prefix := fmt.Sprintf("[%d] ", v.ID)
	log.Printf(prefix+format, args...)
}

func (v *Vehicle) UpdatePosition(val Place) {
	if v.position != nil {
		v.position.Leave()
	}
	val.Take(v)
	v.position = val
}

//NextPlace returns the vehicle's new position
func (v *Vehicle) NextPlace() Place {
	return v.position
}

func (v *Vehicle) ETA() float64 {
	return 0.0
}

/*
DoRound - simulate vehicle's behaviour for a round
*/
func (v *Vehicle) DoRound() error {
	time.Sleep(time.Microsecond * v.position.TravelTime(v.MaxSpeed))
	v.UpdatePosition(v.NextPlace())
	return nil
}

/*
Start - begin simulation of this vehicle
*/
func (v *Vehicle) Start(controller <-chan interface{}, queue chan Event, wg *sync.WaitGroup) {
	v.logf("[info] Starting")
	v.logf("[info] Initial position: %s", v.position)
	defer wg.Done()
	for {
		select {
		case cmd := <-controller:
			v.logf("[comm] Received: %s", cmd)
			if cmd == "quit" {
				v.logf("[cmd] %s - stopping simulation", cmd)
				return
			}
		}
		v.DoRound()
	}
}

func vehicleFromJSON(rawVehicle map[string]*json.RawMessage, nodes []*Node) *Vehicle {
	var routeIDs []int64
	var vehicle Vehicle
	json.Unmarshal(*rawVehicle["id"], &vehicle.ID)
	json.Unmarshal(*rawVehicle["maxSpeed"], &vehicle.MaxSpeed)
	json.Unmarshal(*rawVehicle["capacity"], &vehicle.Capacity)
	json.Unmarshal(*rawVehicle["route"], &routeIDs)
	for _, nodeID := range routeIDs {
		vehicle.Route = append(vehicle.Route, nodes[nodeID-1])
	}
	return &vehicle
}
