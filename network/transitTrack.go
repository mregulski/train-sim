package network

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"math"
)

// TransitTrack is a Track used for moving around
type TransitTrack struct {
	A             *Node       // one end of the track
	B             *Node       // the other end of the track
	Length        float64     // length in [units]
	MaxSpeed      float64     // speed limit for Vehicles using this Track
	id            string      // unique identifier of this track
	activeVehicle *Vehicle    // Vehicle currrently travelling on this Track
	m             *sync.Mutex // for synchronization
}

// EndPoints implements Track.EndPoints
// return Nodes connected by this Track (in arbitrary order)
func (t *TransitTrack) EndPoints() [2]*Node {
	return [2]*Node{t.A, t.B}
}

// ID - implements Track.ID
func (t *TransitTrack) ID() string { return t.id }

//IsAvailable implements Place.IsAvailable
//Return true if this track has no vehicle on it, false otherwise
func (t *TransitTrack) IsAvailable() bool {
	t.m.Lock()
	defer t.m.Unlock()
	return t.activeVehicle == nil
}

//User returns the Vehicle currently travelling over this Track
//Implements Place.User
func (t *TransitTrack) User() *Vehicle {
	t.m.Lock()
	defer t.m.Unlock()
	return t.activeVehicle
}

func (t *TransitTrack) Leave() {
	t.m.Lock()
	defer t.m.Unlock()
	t.activeVehicle = nil
}

func (t *TransitTrack) Take(vehicle *Vehicle) {
	t.m.Lock()
	defer t.m.Unlock()
	t.activeVehicle = vehicle
}

func (t *TransitTrack) TravelTime(speed float64) time.Duration {
	speed = math.Max(t.MaxSpeed, speed)
	return time.Duration(int(t.Length/speed))
}

func (t *TransitTrack) String() string {
	return fmt.Sprintf("{%s Length: %+v MaxSpeed: %+v}",
		t.id, t.Length, t.MaxSpeed)
}

func transitTrackFromJSON(rawTrack map[string]*json.RawMessage, nodes []*Node) Track {
	var a int
	var b int
	var track TransitTrack
	json.Unmarshal(*rawTrack["a"], &a)
	json.Unmarshal(*rawTrack["b"], &b)
	json.Unmarshal(*rawTrack["length"], &track.Length)
	json.Unmarshal(*rawTrack["maxSpeed"], &track.MaxSpeed)
	json.Unmarshal(*rawTrack["id"], &track.id)
	track.A = nodes[a-1]
	track.B = nodes[b-1]
	track.A.Tracks[trackKey(b)] = append(track.A.Tracks[trackKey(b)], &track)
	track.B.Tracks[trackKey(a)] = append(track.B.Tracks[trackKey(a)], &track)
	track.m = &sync.Mutex{}
	return &track
}
