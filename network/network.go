package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

// Network - complete simulation network
type Network struct {
	Nodes    []*Node
	Stations []*Station
	Tracks   map[trackKey2D][]Track
	Vehicles []*Vehicle
}

type Place interface {
	IsAvailable() bool
	User() *Vehicle
	Leave()
	Take(vehicle *Vehicle)
	TravelTime(speed float64) time.Duration
}

func (network *Network) String() string {
	return fmt.Sprintf("{\n Nodes: %+v,\n Tracks: %+v,\n Stations: %+v,\n Vehicles: %+v\n}", network.Nodes, network.Tracks, network.Stations, network.Vehicles)
}

func (network *Network) GetStation(node *Node) *Station {
	for _, station := range network.Stations {
		if station.A == node || station.B == node {
			return station
		}
	}
	return nil
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
	for _, node := range network.Nodes {
		node.Tracks = make(map[trackKey][]Track)
		node.m = &sync.Mutex{}
	}
	var rawTracks []map[string]*json.RawMessage
	if err := json.Unmarshal(*x["tracks"], &rawTracks); err != nil {
		log.Fatal(err.Error())
		return err
	}
	network.Tracks = make(map[trackKey2D][]Track)
	for _, rawTrack := range rawTracks {
		var track Track
		if len(rawTrack) == 4 {
			track = waitTrackFromJSON(rawTrack, network.Nodes)
		} else if len(rawTrack) == 5 {
			track = transitTrackFromJSON(rawTrack, network.Nodes)
		}
		ends := track.EndPoints()
		network.Tracks[trackKey2D{ends[0].ID, ends[1].ID}] =
			append(network.Tracks[trackKey2D{ends[0].ID, ends[1].ID}], track)
		network.Tracks[trackKey2D{ends[1].ID, ends[0].ID}] =
			append(network.Tracks[trackKey2D{ends[0].ID, ends[1].ID}], track)
	}

	var rawStations []map[string]*json.RawMessage
	if err := json.Unmarshal(*x["stations"], &rawStations); err != nil {
		log.Fatal(err.Error())
		return err
	}
	for _, rawStation := range rawStations {
		network.Stations = append(network.Stations, stationFromJSON(rawStation, network.Nodes))
	}

	var rawVehicles []map[string]*json.RawMessage
	if err := json.Unmarshal(*x["vehicles"], &rawVehicles); err != nil {
		log.Fatal(err.Error())
		return err
	}

	for _, rawVehicle := range rawVehicles {
		network.Vehicles = append(network.Vehicles,
			vehicleFromJSON(rawVehicle, network.Nodes))
	}
	return nil
}
