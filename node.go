package main

import (
	"fmt"
	"encoding/json"
)

// Station - named node for the routes
// Consists of 2 switchers connected by WaitTracks
type Station struct {
	A    *Node  `json:"a"`
	B    *Node  `json:"b"`
	Name string `json:"name"`
}

// Node is the basic node of the network
type Node struct {
	ID       int64   `json:"id"`
	WaitTime float64 `json:"waitTime"`
	Tracks   []Track `json:"-"`
}

func (node *Node) String() string {
	return fmt.Sprintf("%+v", *node)
}

func (s Station) String() string {
	return fmt.Sprintf("{A: <%d>, B: <%d>, Name: '%s'}", s.A.ID, s.B.ID, s.Name)
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
	return &station
}