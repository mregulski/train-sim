package network

import (
	"container/heap"
	"math"
)

type item struct {
	previous   Location
	position   Location
	travelTime float64
	index      int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].travelTime < pq[j].travelTime
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) update(item *item, travelTime float64) {
	// item.position = position
	item.travelTime = travelTime
	heap.Fix(pq, item.index)
}

func makeQueue(graph *Graph, blacklist []Location) (priorityQueue, map[Location]*item) {
	queue := make(priorityQueue, 0)
	mapping := make(map[Location]*item)
	for _, junction := range graph.Junctions {
		entry := &item{
			previous:   nil,
			position:   junction,
			travelTime: math.Inf(1),
			index:      len(queue)}
		blacklisted := false
		for _, v := range blacklist {
			if v == junction {
				blacklisted = true
				break
			}
		}
		if !blacklisted {
			queue = append(queue, entry)
			mapping[junction] = entry
		}
	}
	for _, track := range graph.Tracks() {
		entry := &item{
			previous:   nil,
			position:   track,
			travelTime: math.Inf(1),
			index:      len(queue)}
		blacklisted := false
		for _, v := range blacklist {
			if v == track {
				blacklisted = true
				break
			}
		}
		if !blacklisted {
			queue = append(queue, entry)
			mapping[track] = entry
		}
	}
	heap.Init(&queue)
	return queue, mapping
}

func contains(list []interface{}, val interface{}) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
