package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"
)

type trackKeyType struct {
	A, B int
}

type report struct {
	delta int
	key   string
}

// Graph is the entire simulated network
type Graph struct {
	Config        *graphConfig
	Junctions     []*Junction
	tracks        map[trackKeyType][]Track
	Stations      map[string]*Station
	StationLookup map[int]*Station
	Vehicles      []Vehicle
	Emergency     chan emergency
	emergencyCtr  chan report
}

// graphConfig stores general configuration settings of the simulated network
type graphConfig struct {
	TimeScale   float64 // number of milliseconds per simulated hour
	RepairTime  float64 // in hours
	FailureRate float64 // probability of a network element failure per hour
	Tasks       taskConfig
}

type requestHandler interface {
	getRequestChannel() chan<- request
	getRWRequestChannel() chan request
}

/*
Start beging the simulation
*/
func (graph *Graph) Start() {
	rand.Seed(time.Now().UnixNano())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go graph.statsHandler()

	for _, junction := range graph.Junctions {
		wg.Add(1)
		go Handle(junction, graph)
	}

	for _, track := range graph.Tracks() {
		wg.Add(1)
		go Handle(track, graph)
	}
	// for i := 0; i < len(graph.Vehicles); i++ {
	// 	wg.Add(1)
	// 	go graph.Vehicles[i].Handle(graph)
	// }
	graph.Vehicles[5].Handle(graph)

	for _, station := range graph.Stations {
		wg.Add(1)
		go station.Handle(graph)
	}
	wg.Wait()
}

/*
StationWith returns a Station that contains junction j
*/
func (graph *Graph) StationWith(j *Junction) Station {
	return *graph.StationLookup[j.ID]
}

/*
Tracks provide a slice of all unique tracks in the graph
*/
func (graph *Graph) Tracks() []Track {
	visited := make(map[trackKeyType]bool)
	uniqueTracks := []Track{}
	for k, entry := range graph.tracks {
		if (visited[trackKeyType{k.A, k.B}] || visited[trackKeyType{k.B, k.A}]) {
			continue
		}
		for _, track := range entry {
			uniqueTracks = append(uniqueTracks, track)
		}
		visited[k] = true
	}
	return uniqueTracks
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

// generateFailures randomly treis to generate a failure every hour, until it succeeds
func (graph *Graph) generateFailures(accident chan<- bool) {
	ticker := time.NewTicker(graph.scaledTime(1.0))
	<-time.After(graph.scaledTime(2.0))
	for {
		<-ticker.C
		if rand.Float64() < graph.Config.FailureRate {
			accident <- true
			return
		}
	}
}

func (graph *Graph) generateTasks(tasks chan<- task, stop <-chan bool) {
	ticker := time.NewTicker(graph.scaledTime(1.0))
	<-time.After(graph.scaledTime(2.0))
	for {

		<-ticker.C
		if rand.Float64() < graph.Config.Tasks.Rate {
			tasks <- graph.Config.Tasks.randomTask()
		}

	}
}

func (graph *Graph) statsHandler() {
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

func (graph *Graph) UnmarshalJSON(s []byte) error {

	var raw map[string]json.RawMessage
	json.Unmarshal(s, &raw)
	if err := json.Unmarshal(raw["config"], &graph.Config); err != nil {
		log.Panicln("Unable to unmarshal config data: ", err)
	}
	graph.Emergency = make(chan emergency)
	graph.StationLookup = make(map[int]*Station)

	graph.loadJunctions(raw["junctions"])
	graph.loadTracks(raw["tracks"])
	graph.loadStations(raw["stations"])
	graph.loadVehicles(raw["vehicles"])

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
	graph.tracks = make(map[trackKeyType][]Track)
	for _, rawTrack := range rawTracks {
		track := trackFromJSON(rawTrack, graph.Junctions, graph.Config)
		a := track.A().ID
		b := track.B().ID
		track.A().Tracks[b] = append(track.A().Tracks[b], track)
		track.B().Tracks[a] = append(track.B().Tracks[a], track)
		graph.tracks[trackKeyType{a, b}] = append(graph.tracks[trackKeyType{a, b}], track)
		graph.tracks[trackKeyType{b, a}] = append(graph.tracks[trackKeyType{b, a}], track)
	}
}

func (graph *Graph) loadStations(raw json.RawMessage) {
	var rawStations []map[string]*json.RawMessage
	if err := json.Unmarshal(raw, &rawStations); err != nil {
		log.Panicln("Unable to unmarshal station data: ", err)
	}
	graph.Stations = make(map[string]*Station)
	for _, rawStation := range rawStations {
		station := stationFromJSON(rawStation, graph.Junctions)
		graph.Stations[station.name] = station
		graph.StationLookup[station.A.ID] = station
		graph.StationLookup[station.B.ID] = station
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
