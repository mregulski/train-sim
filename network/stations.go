package network

import (
	"encoding/json"
	"fmt"
	"log"
)

// Station is a pair of Junctions connected by WaitTracks
type Station struct {
	A      *Junction
	B      *Junction
	Trains map[Vehicle]struct{}
	name   string
}


/*
Handle manages task creation and execution at the Station
*/
func (s *Station) Handle(ctx *Graph) {
	tasks := make(chan task)
	stop := make(chan bool)
	go ctx.generateTasks(tasks, stop)
	for {
		task := <-tasks
		log.Printf("\n\nnew task: %v\n\n", task)
	}
}

func (s *Station) pickWaitTrack() *WaitTrack {
	return chooseTrack(s.A.Tracks[s.B.ID]).(*WaitTrack)
}

func (s *Station) waitTracks() []Track {
	return s.A.Tracks[s.B.ID]
}

func (s *Station) getRouteTo(target Station) (*Junction, []Track) {
	choices, ok := s.A.Tracks[target.A.ID]
	junction := s.A
	if !ok {
		choices, ok = s.A.Tracks[target.B.ID]
	}
	if !ok {
		choices, ok = s.B.Tracks[target.A.ID]
		junction = s.B
	}
	if !ok {
		choices, ok = s.B.Tracks[target.B.ID]
	}
	if !ok {
		log.Panicf("impossible route: no track exists between %s and %s", s.name, target.name)
	}
	return junction, choices
}

func (s *Station) String() string {
	return fmt.Sprintf("Station{name: %s, A: %d, B: %d}", s.name, s.A.ID, s.B.ID)
}

func commonTrains(a Station, b Station) []Vehicle {
	var common []Vehicle
	for train := range a.Trains {
		if _, ok := b.Trains[train]; ok {
			common = append(common, train)
		}
	}
	return common
}

func stationFromJSON(raw map[string]*json.RawMessage, junctions []*Junction) *Station {
	var station Station
	var a int
	var b int
	json.Unmarshal(*raw["a"], &a)
	json.Unmarshal(*raw["b"], &b)
	json.Unmarshal(*raw["name"], &station.name)
	station.A = junctions[a-1]
	station.B = junctions[b-1]
	station.Trains = make(map[Vehicle]struct{})
	return &station
}
