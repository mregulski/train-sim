package network

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
	"fmt"
)

type TrackKeyType struct {
	A, B int
}

type report struct {
	delta int
	key string
}
// Graph is the entire simulated network
type Graph struct {
	Config    *GraphConfig
	Junctions []*Junction
	Tracks    map[TrackKeyType][]Track
	Stations  []*Station
	Vehicles  []Vehicle
	Emergency chan emergency
	emergencyCtr chan report

}

type requestHandler interface {
	GetRequestChannel() chan<- request
	GetRWRequestChannel() chan request
}

// GraphConfig stores general configuration settings of the simulated network
type GraphConfig struct {
	TimeScale   float64 // number of milliseconds per simulated hour
	RepairTime  float64 // in hours
	FailureRate float64 // probability of a network element failure per hour
}

func (graph *Graph) scaledTime(hours float64) time.Duration {
	return time.Millisecond * time.Duration(graph.Config.TimeScale*hours)
}

func (graph *Graph) waitTime() time.Duration {
	return graph.scaledTime((float64(rand.Intn(30)) + 10.0) / 60)
}

func (graph *Graph) repairTime() time.Duration {
	return graph.scaledTime(graph.Config.RepairTime)
}

// generateFailures randomly generates a failure event once an hour until it receives
// a signal on stop channel
func (graph *Graph) generateFailures(accident chan<- request, stop <-chan bool) {
	ticker := time.NewTicker(graph.scaledTime(1.0))
	<-time.After(graph.scaledTime(2.0))
	for {
		select {
		case <-ticker.C:
			if rand.Float64() < graph.Config.FailureRate {
				accident <- request{nil, -1, fail}
			}
		case <-stop:
			return
		}
	}
}

func (graph *Graph) InfoHandler() {
	graph.emergencyCtr = make(chan report)
	status := make(map[string]struct{})
	activeEmergencies := 0
	for {
		report := <-graph.emergencyCtr
		activeEmergencies += report.delta
		if report.delta > 0 {
			status[report.key] = struct{}{}
		} else if report.delta < 0 {
			delete(status, report.key)
		}
		emergencies := ""
		for k := range status {
			emergencies += fmt.Sprintf("%19s %s\n", "", k)
		}
		log.Printf("[Network] %d active emergencies: \n%s", activeEmergencies, emergencies)
	}
}


// LoadGraph loads a Graph description from a JSON file
func LoadGraph(filename string) (*Graph, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	graph := Graph{}
	if err = json.Unmarshal(raw, &graph); err != nil {
		return nil, err
	}
	return &graph, nil
}

func (graph *Graph) allTracks() []Track {
	list := make([]Track, 0)
	for _, tracks := range graph.Tracks {
		for _, track := range tracks {
			list = append(list, track)
		}
	}
	return list
}

func (graph *Graph) UnmarshalJSON(s []byte) error {

	var raw map[string]json.RawMessage
	json.Unmarshal(s, &raw)
	if err := json.Unmarshal(raw["config"], &graph.Config); err != nil {
		log.Panicln("Unable to unmarshal config data: ", err)
	}
	graph.loadJunctions(raw["junctions"])
	graph.loadTracks(raw["tracks"])
	graph.loadStations(raw["stations"])
	graph.loadVehicles(raw["vehicles"])
	graph.Emergency = make(chan emergency)
	return nil
}

func (graph *Graph) loadJunctions(raw json.RawMessage) {
	var rawJunctions []map[string]*json.RawMessage
	if err := json.Unmarshal(raw, &rawJunctions); err != nil {
		log.Panicln("Unable to unmarshal junction data: ", err)
	}
	for _, rawJunction := range rawJunctions {
		graph.Junctions = append(graph.Junctions, junctionFromJSON(rawJunction, graph.Config))

	}
}

func (graph *Graph) loadTracks(raw json.RawMessage) {
	var rawTracks []map[string]*json.RawMessage
	if err := json.Unmarshal(raw, &rawTracks); err != nil {
		log.Panicln("Unable to unmarshal track data: ", err)
	}
	graph.Tracks = make(map[TrackKeyType][]Track)
	for _, rawTrack := range rawTracks {
		track := trackFromJSON(rawTrack, graph.Junctions, graph.Config)
		a := track.A().ID
		b := track.B().ID
		track.A().Tracks[b] = append(track.A().Tracks[b], track)
		track.B().Tracks[a] = append(track.B().Tracks[a], track)
		graph.Tracks[TrackKeyType{a, b}] = append(graph.Tracks[TrackKeyType{a, b}], track)
		graph.Tracks[TrackKeyType{b, a}] = append(graph.Tracks[TrackKeyType{a, b}], track)
	}
}

func (graph *Graph) loadStations(raw json.RawMessage) {
	var rawStations []map[string]*json.RawMessage
	if err := json.Unmarshal(raw, &rawStations); err != nil {
		log.Panicln("Unable to unmarshal station data: ", err)
	}
	for _, rawStation := range rawStations {
		graph.Stations = append(graph.Stations, stationFromJSON(rawStation, graph.Junctions))
	}
}

func (graph *Graph) loadVehicles(raw json.RawMessage) {
	var rawVehicles []map[string]*json.RawMessage
	if err := json.Unmarshal(raw, &rawVehicles); err != nil {
		log.Panicln("Unable to unmarshal vehicle data: ", err)
	}
	for _, rawVehicle := range rawVehicles {
		graph.Vehicles = append(graph.Vehicles, vehicleFromJSON(rawVehicle, graph))
	}
}
