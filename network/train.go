package network

import (
	"encoding/json"
	"fmt"
	"time"
)

// Train is a basic vehicle travelling through the network along a predefined route
type Train struct {
	baseVehicle
	Route    []*Station
	capacity int
	accident chan bool
	requests chan request
}

func (t *Train) String() string {
	route := ""
	for i, station := range t.Route {
		route += station.name
		if i < len(t.Route)-1 {
			route += ", "
		}

	}
	return fmt.Sprintf("Train{maxSpeed: %f, capacity: %d, route: %s}", t.maxSpeed, t.capacity, route)
}

/*
Handle manages the trains behaviour and communications
*/
func (t *Train) Handle(ctx *Graph) {
	var curLocation Location
	var stationIdx = 0
	var curStation = t.Route[stationIdx]

	curLocation = t.travelToOneOf(curStation.waitTracks(), nil, ctx)
	t.logf("Starting at %s", curLocation.Name())
	fails := make(chan bool)
	go ctx.generateFailures(fails)
	laps := 0
	for {
		stationIdx = t.nextStationIdx(stationIdx)
		nextStation := t.Route[stationIdx]
		t.logf("Next station: %s", nextStation.name)
		start, trackChoices := curStation.getRouteTo(*nextStation)
		curLocation = t.travelTo(start, curLocation, false, ctx)
		t.maybeFailAndRecover(curLocation, fails, ctx)
		curLocation = t.travelToOneOf(trackChoices, curLocation, ctx)
		t.maybeFailAndRecover(curLocation, fails, ctx)
		curLocation = t.travelTo(curLocation.(Track).oppositeEnd(start), curLocation, false, ctx)
		t.maybeFailAndRecover(curLocation, fails, ctx)
		curLocation = t.travelTo(nextStation.pickWaitTrack(), curLocation, false, ctx)
		t.maybeFailAndRecover(curLocation, fails, ctx)
		curStation = nextStation

		t.logf("Arrived at station %s", curStation.name)
		if curStation == t.Route[0] {
			laps++
			t.logf("Route completed (%d times so far)", laps)
		}
	}

}

/*
getRequestChannel - implements requestHandler.getRequestChannel
*/
func (t *Train) getRequestChannel() chan<- request {
	return t.requests
}

/*
getRWRequestChannel - implements requestHandler.getRWRequestChannel
*/
func (t *Train) getRWRequestChannel() chan request {
	return t.requests
}

func (t *Train) travelTo(location Location, from Location, once bool, ctx *Graph) Location {
	delay := ctx.waitTime()

	// enter the new location
	t.logf("Requesting entry: %s", location.Name())
	for !t.request(location, take) {
		t.logf("%s - entry denied", location.Name())
		if once {
			return nil
		}
		// check failure reason
		if failing := !t.request(location, check); failing {
			t.logf("Destination offline, retrying after %v", delay*2)
			<-time.After(delay * 2)
		} else {
			t.logf("Destination occupied, retrying after %v", delay)
			<-time.After(delay)
		}
		continue
	}
	t.logf("Arrived at %s", location.Name())
	if from != nil { // => we're not doing the initial setup
		// free the previous one
		for !t.request(from, free) {
			t.logf("Unable to leave previous location: %s - retrying after %v", from, delay)
			<-time.After(delay)
		}
		t.logf("Left %s", from.Name())
	}

	// simulate travel through the new location
	travelTime := ctx.scaledTime(location.TravelTime(t.maxSpeed))
	t.logf("Traversing %s, ETA: %v", location.Name(), travelTime)
	<-time.After(travelTime)

	return location
}

func (t *Train) travelToOneOf(trackChoices []Track, from Location, ctx *Graph) Location {
	chosen := chooseTrack(trackChoices)
	dst := t.travelTo(chosen, from, true, ctx)
	for dst == nil {
		<-time.After(ctx.waitTime())
		chosen = chooseTrack(trackChoices)
		t.logf("Trying another track: %s", chosen.Name())
		dst = t.travelTo(chosen, from, true, ctx)
	}
	return dst
}

func (t *Train) nextStationIdx(idx int) int {
	return (idx + 1) % len(t.Route)
}

func (t *Train) maybeFailAndRecover(curLocation Location, fails chan bool, ctx *Graph) {
	select {
	case <-fails:
		ctx.emergencyCtr <- report{delta: 1, key: fmt.Sprintf("Train #%d", t.id)}
		ctx.Emergency <- emergency{curLocation, t}
		t.awaitRepair()
		go ctx.generateFailures(fails)
	default:
		// hurray, no train crash! (for now)
	}
}

/**
AwaitRepair causes train to ignore all requests and wait in its current
location until it is repaired
*/
func (t *Train) awaitRepair() {
	var req request
	req = <-t.requests
	for req.kind != repairStart {
		req.c <- false
		req = <-t.requests
	}
	req.c <- true
	req = <-t.requests
	for req.kind != repairDone {
		req.c <- false
		req = <-t.requests
	}
	req.c <- true

}

func trainFromJSON(raw map[string]*json.RawMessage, junctions []*Junction,
	context *Graph) *Train {

	var stationNames []string
	var train Train
	json.Unmarshal(*raw["id"], &train.id)
	json.Unmarshal(*raw["maxSpeed"], &train.maxSpeed)
	json.Unmarshal(*raw["capacity"], &train.capacity)
	json.Unmarshal(*raw["route"], &stationNames)
	for _, stationName := range stationNames {
		station := context.Stations[stationName]
		train.Route = append(train.Route, station)
		station.Trains[&train] = struct{}{}
	}
	train.comm = make(chan bool)
	return &train
}
