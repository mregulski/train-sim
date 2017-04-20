package network

import (
	"encoding/json"
	"fmt"
	"sync"
)

// WaitTrack is a Track used for waiting for passengers
type WaitTrack struct {
	// one end of this Track
	A *Node
	// the other end of this Track
	B *Node
	// minimum amount of time a Vehicle must wait for on this Track
	WaitTime float64
	// unique identifier of this Track
	id string
	// Vehicle currently waiting on this Track
	activeVehicle *Vehicle
	// For synchronization
	m *sync.Mutex
}

// EndPoints - implements Track.EndPoints
// return Nodes connected by this Track (in arbitrary order)
func (t *WaitTrack) EndPoints() [2]*Node {
	return [2]*Node{t.A, t.B}
}

// ID - implements Track.ID
func (t *WaitTrack) ID() string { return t.id }

func (t *WaitTrack) IsAvailable() bool {
	t.m.Lock()
	defer t.m.Unlock()
	return t.activeVehicle == nil
}

func (t *WaitTrack) User() *Vehicle {
	t.m.Lock()
	defer t.m.Unlock()
	return t.activeVehicle
}

func (t *WaitTrack) Leave() {
	t.m.Lock()
	defer t.m.Unlock()
	t.activeVehicle = nil
}

func (t *WaitTrack) Take(vehicle *Vehicle) bool {
	t.m.Lock()
	defer t.m.Unlock()
	if (t.activeVehicle == nil) {
		t.activeVehicle = vehicle
		return true
	}
	return t.activeVehicle == vehicle
}

func (t *WaitTrack) TravelTime(speed float64) int {
	return int(t.WaitTime)
}

func (t *WaitTrack) Name() string {
	return fmt.Sprintf("<%s>", t.id)
}

func (t *WaitTrack) String() string {
	return fmt.Sprintf("{%s WaitTime: %+v}",
		t.id, t.WaitTime)
}

func waitTrackFromJSON(rawTrack map[string]*json.RawMessage, nodes []*Node) Track {
	var a int
	var b int
	var track WaitTrack
	json.Unmarshal(*rawTrack["a"], &a)
	json.Unmarshal(*rawTrack["b"], &b)
	json.Unmarshal(*rawTrack["waitTime"], &track.WaitTime)
	json.Unmarshal(*rawTrack["id"], &track.id)
	track.A = nodes[a-1]
	track.B = nodes[b-1]
	track.A.Tracks[trackKey(b)] = append(track.A.Tracks[trackKey(b)], &track)
	track.B.Tracks[trackKey(a)] = append(track.B.Tracks[trackKey(a)], &track)
	track.m = &sync.Mutex{}
	return &track
}
