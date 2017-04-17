package main

import (
	"fmt"
	"encoding/json"
	"log"
	"io/ioutil"
)

// Network - complete simulation network
type Network struct {
	Nodes    []*Node    `json:"nodes"`
	Stations []*Station `json:"stations"`
	Tracks   []Track    `json:"tracks"`
	Vehicles []*Vehicle `json:"vehicles"`
}

func (network Network) String() string {
	return fmt.Sprintf("{\n Nodes: %+v,\n Tracks: %+v,\n Stations: %+v,\n Vehicles: %+v\n}", network.Nodes, network.Tracks, network.Stations, network.Vehicles)
}

func (network *Network) LoadFromJSONFile(filename string) error {
	raw, err := ioutil.ReadFile("network.json")
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	err = json.Unmarshal(raw, network)
	return err
}

/*
UnmarshalJSON - implements Unmarshaler.UnmarshalJSON
Cannot use default implementation because network nodes are referenced by id
and we want them to be referenced directly in the Network object
*/ 
func (network *Network) UnmarshalJSON(s []byte) error {
	var x map[string]*json.RawMessage
	json.Unmarshal(s, &x)
	if err := json.Unmarshal(*x["nodes"], &network.Nodes); err != nil {
		log.Fatal(err.Error())
		return err
	}
	var rawTracks []map[string]*json.RawMessage
	if err := json.Unmarshal(*x["tracks"], &rawTracks); err != nil {
		log.Fatal(err.Error())
		return err
	}
	for _, rawTrack := range rawTracks {
		var track Track
		if len(rawTrack) == 3 {
			track = waitTrackFromJSON(rawTrack, network.Nodes)
		} else if len(rawTrack) == 4 {
			track = transitTrackFromJSON(rawTrack, network.Nodes)
		}
		network.Tracks = append(network.Tracks, track)
	}

	var rawStations []map[string]*json.RawMessage
	if err := json.Unmarshal(*x["stations"], &rawStations); err != nil {
		log.Fatal(err.Error())
		return err
	}
	for _, rawStation := range rawStations {
		network.Stations = append(network.Stations, stationFromJSON(rawStation, network.Nodes))
	}

	if err := json.Unmarshal(*x["vehicles"], &network.Vehicles); err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}
