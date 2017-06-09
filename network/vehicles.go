package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Vehicle interface {
	// MaxSpeed is vehicle's max possible speed, in km/h
	MaxSpeed() float64
	// Handle implements the vehicle's logic and communication with network elements
	Handle(graph *Graph)
}

type BaseVehicle struct {
	ID       int
	maxSpeed float64 // in km/h
	comm     chan bool
}

// Train is a basic vehicle travelling through the network along a predefined route
type Train struct {
	BaseVehicle
	Route    []*Junction
	capacity int
	accident chan bool
	requests chan request
	routeIdx int
}

func (v *BaseVehicle) MaxSpeed() float64 {
	return v.maxSpeed
}

func (t *Train) GetRequestChannel() chan<- request {
	return t.requests
}

func (t *Train) GetRWRequestChannel() chan request {
	return t.requests
}

func (v *BaseVehicle) request(target requestHandler, req requestType) bool {
	target.GetRequestChannel() <- request{v.comm, v.ID, req}
	return <-v.comm
}

func (t *Train) Handle(context *Graph) {
	var position Location
	stop := make(chan bool)
	go context.generateFailures(t.requests, stop)
	laps := 0
	failure := false
	for {
		select {
		case <-t.requests:
		 	t.logf("FAILURE, requesting help to %s", position.Name())
			stop <- true
			context.Emergency <- emergency{location: position, handler: t}
			failure = true
			// req.re
		default:
			//do notihng
		}
		if failure {
			var req request
			req =<- t.requests
			for req.kind != repairStart {
				req.c <- false
				req = <-t.requests
			}
			req.c <- true
			req =<- t.requests
			for req.kind != repairDone {
				req.c <- false
				req = <-t.requests
			}
			req.c <- true
			failure = false
		} else {
			var next Location
			if position == nil {
				next = t.initialPosition()
				t.logf("Starting position: %s", next.Name())
			} else {
				next = t.next(position, context.waitTime)
				// t.logf("Next destination: %s", next.Name())
			}
			for {
				t.logf("Requesting entry: %s", next.Name())
				ok := t.request(next, take)
				if !ok {
					t.logf("Entry denied")
					delay := context.waitTime()
					if failing := t.request(next, check); failing {
						t.logf("Destination offline, retrying after %v", delay)
					} else {
						t.logf("Destination busy (%d), retrying after %v", next.Occupant(), delay)
					}

					<-time.After(delay)
					continue
				} else {
					break
				}
			}
			if position != nil {
				for !t.request(position, free) {
					t.logf("Unable to leave current location: %s", position)
					<-time.After(context.waitTime())

				}
			}
			position = next
			// get through current poistion
			travelTime := context.scaledTime(position.TravelTime(t.maxSpeed))
			t.logf("Travelling %s, eta: %v", position.Name(), travelTime)
			<-time.After(travelTime)
			if position == t.Route[0] {
				laps++
				t.logf("Vehicle has completed its route (%d times so far)", laps)
			}
		}
	}
}

func (v *BaseVehicle) logf(format string, args ...interface{}) {
	prefix := fmt.Sprintf("\x1b[3"+strconv.Itoa(v.ID%9)+"m[Train #%d] ", v.ID)
	log.Printf(prefix+format+"\x1b[39m", args...)
}

func (t *Train) initialPosition() Location {
	return t.next(t.Route[0], nil)
}

func (t *Train) next(position Location, delay func() time.Duration) Location {
	var destination Location
	switch pos := position.(type) {
	case Track:
		t.routeIdx = t.nextIdx()
		destination = t.Route[t.routeIdx]
	case *Junction:

		target := t.Route[t.nextIdx()]
		tracks := pos.Tracks[target.ID]
		if len(tracks) == 0 {
			t.logf("Malformed route: No tracks found between %s and %s\n",
				pos.Name(), target.Name())
			panic(errors.New("Malformed route. Check logs for details"))
		}
		for {
			if destination = findFreeTrack(tracks); destination != nil {
				break
			} else if delay != nil {
				wait := delay()
				t.logf("No free track between %s and %s, retrying after %v\n",
					pos.Name(), target.Name(), wait)
				<-time.After(wait)
			}
		}
	}
	return destination
}

func (t *Train) nextIdx() int {
	return (t.routeIdx + 1) % len(t.Route)
}

func vehicleFromJSON(raw map[string]*json.RawMessage, graph *Graph) Vehicle {

	var kind string
	var vehicle Vehicle
	json.Unmarshal(*raw["type"], &kind)
	switch kind {
	case "train":
		vehicle = trainFromJSON(raw, graph.Junctions, graph.Config)
	case "repair":
		vehicle = repairFromJSON(raw, graph.allTracks(), graph.Config)
	}
	return vehicle
}

func trainFromJSON(raw map[string]*json.RawMessage, junctions []*Junction,
	config *GraphConfig) *Train {

	var routeIDs []int
	var vehicle Train
    log.Printf("wtf: %s", vehicle)
	json.Unmarshal(*raw["id"], &vehicle.ID)
	json.Unmarshal(*raw["maxSpeed"], &vehicle.maxSpeed)
	json.Unmarshal(*raw["capacity"], &vehicle.capacity)
	json.Unmarshal(*raw["route"], &routeIDs)
	for _, nodeID := range routeIDs {
		vehicle.Route = append(vehicle.Route, junctions[nodeID-1])
	}
	vehicle.comm = make(chan bool)
	// vehicle.config = config
	return &vehicle
}
