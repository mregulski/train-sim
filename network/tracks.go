package network

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
)

type Track interface {
	Location
	A() *Junction
	B() *Junction
	id() string
}

// baseTrack is a base struct representing a bidirectional track from a to b.
// Should be extended by concrete types.
type baseTrack struct {
	basePosition
	a   *Junction
	b   *Junction
	_id string
}

func (track *baseTrack) A() *Junction {
	return track.a
}

func (track *baseTrack) B() *Junction {
	return track.b
}

func (track *baseTrack) Occupied() bool {
	return track.occupant != -1
}

// Name - implements Position.Name
func (track *baseTrack) Name() string {
	return track._id
}

func (track *baseTrack) id() string {
	return track._id
}

func (track *baseTrack) neighbours() []Location {
	return []Location{track.a, track.b}
}

func findFreeTrack(tracks []Track) Track {
	idx := rand.Intn(len(tracks))
	return tracks[idx]
	// for _, track := range tracks {
	// 	// if !track.Occupied() {
	// 	// 	return track
	// 	// }
	// }
	// return nil
}

// WaitTrack is a Track with constant time of traversal, independent on the Vehicle's speed
type WaitTrack struct {
	baseTrack
	WaitTime float64
}

// TravelTime - implements Position.TravelTime
func (wt *WaitTrack) TravelTime(speed float64) float64 {
	return wt.WaitTime
}

func (wt *WaitTrack) String() string {
	return fmt.Sprintf("WaitTrack{id: %s, A: %d, B: %d, waitTime: %.2f}",
		wt._id, wt.a.ID, wt.b.ID, wt.WaitTime)
}

// TransitTrack is a Track with time of traversal depending on it's length,``
// and Vehicle's speed, possibly limited by the track
type TransitTrack struct {
	baseTrack
	Length   float64
	MaxSpeed float64
}

// TravelTime - implements Position.TravelTime
func (tt *TransitTrack) TravelTime(speed float64) float64 {
	return tt.Length / tt.MaxSpeed
}

func (tt *TransitTrack) String() string {
	return fmt.Sprintf("TransitTrack{id: %s, A: %d, B: %d, length: %.2f, maxSpeed: %.2f}",
		tt._id, tt.a.ID, tt.b.ID, tt.Length, tt.MaxSpeed)
}

func trackFromJSON(raw map[string]*json.RawMessage, junctions []*Junction, config *GraphConfig) Track {
	var kind string
	var track Track
	json.Unmarshal(*raw["type"], &kind)
	switch kind {
	case "transit":
		track = transitTrackFromJSON(raw, junctions, config)
	case "wait":
		track = waitTrackFromJSON(raw, junctions, config)
	default:
		log.Panicln("Unknown track type: ", kind)
	}
	return track
}

func transitTrackFromJSON(raw map[string]*json.RawMessage, junctions []*Junction,
	config *GraphConfig) *TransitTrack {

	var a int
	var b int
	var track TransitTrack
	track.basePosition = basePosition{config, -1, make(chan request), make(chan request)}
	json.Unmarshal(*raw["a"], &a)
	json.Unmarshal(*raw["b"], &b)
	json.Unmarshal(*raw["length"], &track.Length)
	json.Unmarshal(*raw["maxSpeed"], &track.MaxSpeed)
	json.Unmarshal(*raw["id"], &track._id)
	track.a = junctions[a-1]
	track.b = junctions[b-1]
	return &track
}

func waitTrackFromJSON(raw map[string]*json.RawMessage, junctions []*Junction,
	config *GraphConfig) *WaitTrack {

	var a int
	var b int
	var track WaitTrack
	track.basePosition = basePosition{config, -1, make(chan request), make(chan request)}
	json.Unmarshal(*raw["a"], &a)
	json.Unmarshal(*raw["b"], &b)
	json.Unmarshal(*raw["waitTime"], &track.WaitTime)
	json.Unmarshal(*raw["id"], &track._id)
	track.WaitTime /= 60 // minutes in json -> hours
	track.a = junctions[a-1]
	track.b = junctions[b-1]
	return &track
}
