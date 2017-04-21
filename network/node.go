package network

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Station - named node for the routes
// Consists of 2 switchers connected by WaitTracks
type Station struct {
	A    *Node
	B    *Node
	Name string
}

// Node is the basic node of the network
type Node struct {
	ID            int
	WaitTime      float64
	Tracks        map[trackKey][]Track
	m             *sync.Mutex
	activeVehicle *Vehicle
	station       *Station
}

type trackKey2D struct {
	idA int
	idB int
}

type trackKey int

func (n *Node) IsAvailable() bool {
	n.m.Lock()
	defer n.m.Unlock()
	return n.activeVehicle == nil
}

func (n *Node) User() *Vehicle {
	n.m.Lock()
	defer n.m.Unlock()
	return n.activeVehicle
}

func (n *Node) Leave() {
	n.m.Lock()
	defer n.m.Unlock()
	n.activeVehicle = nil
}

func (n *Node) Take(vehicle *Vehicle) bool {
	n.m.Lock()
	defer n.m.Unlock()
	if n.activeVehicle == nil {
		n.activeVehicle = vehicle
		return true
	}
	return n.activeVehicle == vehicle
}

func (n *Node) TravelTime(speed float64) int {
	return int(n.WaitTime / 60)
}

func (n *Node) Name() string {
	if n.station == nil {
		return fmt.Sprintf("<node:%d>", n.ID)
	}
	return fmt.Sprintf("<%s:%d>", n.station.Name, n.ID)
}

func (n *Node) String() string {
	return fmt.Sprintf("%+v", *n)
}

func (s Station) String() string {
	return fmt.Sprintf("{A: <%d>, B: <%d>, Name: '%s'}", s.A.ID, s.B.ID, s.Name)
}

// GetFreeTrack - find first available wait track on the station
// Returns first available track or nil if all are taken
func (s *Station) GetFreeTrack() Track {
	for _, track := range s.A.Tracks[trackKey(s.B.ID)] {
		if track.(Location).IsAvailable() {
			return track
		}
	}
	return nil
}

func stationFromJSON(rawStation map[string]*json.RawMessage, nodes []*Node) *Station {
	var station Station
	var a int
	var b int
	json.Unmarshal(*rawStation["a"], &a)
	json.Unmarshal(*rawStation["b"], &b)
	json.Unmarshal(*rawStation["name"], &station.Name)
	station.A = nodes[a-1]
	station.B = nodes[b-1]
	station.A.station = &station
	station.B.station = &station
	return &station
}
