package main

import (
	"encoding/json"
	"fmt"
)

// Track - a connection between 2 nodes
type Track interface {
	EndPoints() [2]*Node
}

// TransitTrack is a Track used for moving around
type TransitTrack struct {
	A        *Node   `json:"a"`
	B        *Node   `json:"b"`
	Length   float64 `json:"length"`
	MaxSpeed float64 `json:"maxSpeed"`
}

// WaitTrack is a Track used for waiting for passengers
type WaitTrack struct {
	A        *Node   `json:"a"`
	B        *Node   `json:"b"`
	WaitTime float64 `json:"waitTime"`
}

// EndPoints - implements Track.EndPoints
// return Nodes connected by this Track (in arbitrary order)
func (track *TransitTrack) EndPoints() [2]*Node {
	return [2]*Node{track.A, track.B}
}

// EndPoints - implements Track.EndPoints
// return Nodes connected by this Track (in arbitrary order)
func (track *WaitTrack) EndPoints() [2]*Node {
	return [2]*Node{track.A, track.B}
}

func (track *WaitTrack) String() string {
	return fmt.Sprintf("{A: <%d> B: <%d> WaitTime: %+v}", track.A.ID, track.B.ID, track.WaitTime)
}

func (track *TransitTrack) String() string {
	return fmt.Sprintf("{A: <%d> B: <%d> Length: %+v MaxSpeed: %+v}", track.A.ID, track.B.ID, track.Length, track.MaxSpeed)
}

func waitTrackFromJSON(rawTrack map[string]*json.RawMessage, nodes []*Node) Track {
	var a int
	var b int
	var track WaitTrack
	json.Unmarshal(*rawTrack["a"], &a)
	json.Unmarshal(*rawTrack["b"], &b)
	json.Unmarshal(*rawTrack["waitTime"], &track.WaitTime)
	track.A = nodes[a-1]
	track.B = nodes[b-1]
	track.A.Tracks = append(track.A.Tracks, &track)
	track.B.Tracks = append(track.B.Tracks, &track)
	return &track
}

func transitTrackFromJSON(rawTrack map[string]*json.RawMessage, nodes []*Node) Track {
	var a int
	var b int
	var track TransitTrack
	json.Unmarshal(*rawTrack["a"], &a)
	json.Unmarshal(*rawTrack["b"], &b)
	json.Unmarshal(*rawTrack["length"], &track.Length)
	json.Unmarshal(*rawTrack["maxSpeed"], &track.MaxSpeed)
	track.A = nodes[a-1]
	track.B = nodes[b-1]
	track.A.Tracks = append(track.A.Tracks, &track)
	track.B.Tracks = append(track.B.Tracks, &track)
	return &track
}
