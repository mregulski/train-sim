package network

import (
	"encoding/json"
	"fmt"
)

// Junction is a network's vertex
type Junction struct {
	basePosition
	ID       int
	Tracks   map[int][]Track // target junction id -> tracks to that junction
	WaitTime float64
}

/*
Name - implements Location.Name()
*/
func (j *Junction) Name() string {
	return fmt.Sprintf("Junction #%d", j.ID)
}

/*
TravelTime - implements Location.TravelTime().
Returns time, in hours, the junction needs to rotate.
*/
func (j *Junction) TravelTime(speed float64) float64 {
	return j.WaitTime
}

// func (j *Junction) Occupied() bool {
// 	return j.occupant != -1
// }

func (j *Junction) String() string {
	return fmt.Sprintf("Junction{id: %d, waitTime: %.2f}", j.ID, j.WaitTime)
}

func (j *Junction) neighbours() []Location {
	neighbours := make([]Location, 0)
	for _, track := range j.allTracks() {
		neighbours = append(neighbours, track)
	}
	return neighbours
}

func (j *Junction) allTracks() []Track {
	list := make([]Track, 0)
	for _, tracks := range j.Tracks {
		for _, track := range tracks {
			list = append(list, track)
		}
	}
	return list
}

func junctionFromJSON(raw map[string]*json.RawMessage, config *graphConfig) *Junction {
	var junction Junction
	junction.basePosition = basePosition{config, -1, make(chan request), make(chan request)}
	json.Unmarshal(*raw["id"], &junction.ID)
	json.Unmarshal(*raw["waitTime"], &junction.WaitTime)
	junction.WaitTime /= 60 // minutes in json -> hours
	junction.Tracks = make(map[int][]Track)
	return &junction
}
