package network

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type Vehicle interface {
	// MaxSpeed is vehicle's max possible speed, in km/h
	MaxSpeed() float64
	// Handle implements the vehicle's logic and communication with network elements
	Handle(graph *Graph)
	ID() int
}

type baseVehicle struct {
	id       int
	maxSpeed float64 // in km/h
	comm     chan bool
}

func (v *baseVehicle) ID() int {
	return v.id
}

func (v *baseVehicle) MaxSpeed() float64 {
	return v.maxSpeed
}

func (v *baseVehicle) request(target requestHandler, req requestType) bool {
	target.getRequestChannel() <- request{v.comm, v.id, req}
	return <-v.comm
}

func (v *baseVehicle) logf(format string, args ...interface{}) {
	prefix := fmt.Sprintf("\x1b[3"+strconv.Itoa(v.id%9)+"m[Train #%d] ", v.id)
	log.Printf(prefix+format+"\x1b[39m", args...)
}

func vehicleFromJSON(raw map[string]*json.RawMessage, graph *Graph) Vehicle {

	var kind string
	var vehicle Vehicle
	json.Unmarshal(*raw["type"], &kind)
	switch kind {
	case "train":
		vehicle = trainFromJSON(raw, graph.Junctions, graph)
	case "repair":
		vehicle = repairFromJSON(raw, graph.Tracks(), graph.Config)
	}
	return vehicle
}
