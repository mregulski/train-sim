package network

// Track - a connection between 2 nodes
// Embeds Place
type Track interface {
	EndPoints() [2]*Node
	ID() string
	Location
}

type TrackCollection []Track

func (tc TrackCollection) GetFreeTrack() Track {
	for _, track := range tc {
		if track.(Location).IsAvailable() {
			return track
		}
	}
	return nil
}
