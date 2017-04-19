package network

// Track - a connection between 2 nodes
type Track interface {
	EndPoints() [2]*Node
	ID() string
}
