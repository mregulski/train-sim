package network

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"time"
)

type RepairVehicle struct {
	baseVehicle
	Base *WaitTrack
}

func (rv *RepairVehicle) String() string {
	return fmt.Sprintf("Repair{maxSpeed: %f, base: %s}", rv.maxSpeed, rv.Base.Name())
}

func (rv *RepairVehicle) travelTo(location Location, from Location, reserve bool, ctx *Graph) Location {
	success := false
	var start = from
	var blocked Location
	var blacklist = []Location{}
	for !success {
		rv.logf("[Repair] Calculating shortest path to %s", location.Name())
		path := rv.findPath(start, location, ctx, blacklist)
		if len(path) == 0 {
			rv.logf("[Repair] Already close enough for repairs")
			success = true
			continue
		}
		if reserve {
			rv.reserve(path)
		}
		rv.logf("[Repair] Path reserved")

		start, blocked, success = rv.travelByPath(path, ctx)
		if !success {
			rv.logf("[Repair] Path blocked, retrying from %s", start.Name())
			rv.release(path)
			blacklist = []Location{blocked}
		} else {
			blacklist = []Location{}
		}
	}
	return start
}

func (rv *RepairVehicle) Handle(context *Graph) {
	var accidents = make(chan emergency, 10)
	go func(queue chan emergency) {
		for {
			report := <-context.Emergency
			rv.logf("[Repair] Received emergency report from %s", report.location.Name())
			queue <- report
		}
	}(accidents)
	rv.moveTo(rv.Base, nil, context)
	rv.logf("Arrived at base (%s)", rv.Base.Name())
	for {
		accident := <-accidents
		if accident.location == rv.Base {
			rv.repair(accident.location, context)
			continue
		}
		var repairLocation = rv.travelTo(accident.location, rv.Base, true, context)
		rv.logf("[Repair] Arrived at target location %s, beginning repair", accident.location.Name())
		rv.repair(accident.handler, context)
		rv.travelTo(rv.Base, repairLocation, false, context)
		rv.logf("[Repair] Repair done")
	}
}

func (rv *RepairVehicle) travelByPath(path []Location, context *Graph) (Location, Location, bool) {
	var lastLoc = path[0]

	for _, loc := range path {
		if rv.moveTo(loc, lastLoc, context) {
			lastLoc = loc
		} else {
			return lastLoc, loc, false
		}
	}
	return nil, lastLoc, true
}

func (rv *RepairVehicle) reserve(path []Location) {
	rv.logf("[Repair] Reserving path: %v", path)
	for _, place := range path {
		rv.request(place, reserve)
		rv.logf("[Repair] Reserved %s", place.Name())
	}
}

func (rv *RepairVehicle) release(path []Location) {
	rv.logf("[Repair] Reserving path: %v", path)
	for _, place := range path {
		rv.request(place, release)
		rv.logf("[Repair] Reserved %s", place.Name())
	}
}

func (rv *RepairVehicle) repair(target requestHandler, context *Graph) {
	switch target := target.(type) {
	case Location:
		done := false
		for !done {
			rv.request(target, repairStart)
			rv.logf("[Repair] started repairing %s (%v)", target.Name(), context.repairTime())
			<-time.After(context.repairTime())
			done = rv.request(target, repairDone)
		}
		rv.logf("[Repair] %s is back online", target.Name())
		context.emergencyCtr <- report{delta: -1, key: target.Name()}
	case *Train:
		done := false
		for !done {
			rv.request(target, repairStart)
			rv.logf("[Repair] started repairing Train#%d (%v)", target.id, context.repairTime())
			<-time.After(context.repairTime())
			done = rv.request(target, repairDone)
		}
		rv.logf("[Repair] Train#%d is back online", target.id)
		context.emergencyCtr <- report{delta: -1, key: fmt.Sprintf("Train#%d", target.id)}
	}

}

/*
moveTo tries to move rv to pos. Gives up after 5 failed attempts to allow
trying a different path
*/
func (rv *RepairVehicle) moveTo(pos Location, from Location, context *Graph) bool {
	ok := false
	ctr := 0
	for !ok {
		rv.logf("Requesting entry: %s", pos.Name())
		ok = rv.request(pos, take)

		if !rv.request(pos, check) {
			rv.logf("Unexpected emergency in %s, repairing", pos.Name())
			rv.repair(pos, context)
		}

		if !ok {
			ctr++
			delay := context.waitTime()
			rv.logf("Destination occupied, retrying after %v", delay)
			<-time.After(delay)
			if ctr >= 5 {
				return false
			}
		}

	}
	rv.logf("entered %s", pos.Name())
	if from != nil { // => it's not the initial setup
		// todo: sometimes called twice - investigate
		rv.request(from, free)
		rv.request(from, release) // ensure, even if route wasn't actually reserved
		rv.logf("Released %s", pos.Name())
	}
	<-time.After(context.scaledTime(pos.TravelTime(rv.maxSpeed)))
	return true
}

// findPath finds a shortest (by travel time) sequence of positions from repair team's base
//	to the target, using Dijkstra's algorithm
func (rv *RepairVehicle) findPath(source Location, target Location,
	graph *Graph, blackList []Location) []Location {
	rv.logf(">>[Repair] blacklist: %v", blackList)
	queue, items := makeQueue(graph, blackList)
	rv.logf(">>[Repair] created queue")
	rv.logf(">>[Repair] source: %v (%v)", source, items[source])
	queue.update(items[source], 0)
	// heap.Init(&queue)
	for len(queue) > 0 {
		nearest := heap.Pop(&queue).(*item).position
		// rv.logf("[Repair: Path] nearest: %s (%f)", nearest.Name(), items[nearest].travelTime)
		neighbours := nearest.neighbours()
		for _, pos := range neighbours {
			alternative := items[nearest].travelTime + pos.TravelTime(rv.maxSpeed)
			if items[pos] != nil && items[pos].travelTime > alternative {
				queue.update(items[pos], alternative)
				items[pos].previous = nearest
			}
		}
	}
	path := make([]Location, 0)
	last := items[target].position
	rv.logf("[Repair: Path] source: %s", source.Name())
	for last != source {
		rv.logf("[Repair: Path] last: %s", last.Name())
		path = append(path, last)
		last = items[last].previous
	}
	for i := len(path)/2 - 1; i >= 0; i-- {
		opp := len(path) - 1 - i
		path[i], path[opp] = path[opp], path[i]
	}
	return path
}

func repairFromJSON(raw map[string]*json.RawMessage, tracks []Track,
	config *graphConfig) *RepairVehicle {

	var rv RepairVehicle
	var id string

	json.Unmarshal(*raw["id"], &rv.id)
	json.Unmarshal(*raw["maxSpeed"], &rv.maxSpeed)
	json.Unmarshal(*raw["base"], &id)
	for _, track := range tracks {
		if track.id() == id {
			rv.Base = track.(*WaitTrack)
			break
		}
	}
	rv.comm = make(chan bool)
	return &rv
}
