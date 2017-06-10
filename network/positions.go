package network

import (
	"fmt"
	"log"
	// "sync"
)

// Location is an element of the transport network that can be used by a vehicle
type Location interface {
	requestHandler
	// TravelTime returns time in hours required to get through this location at given speed
	TravelTime(speed float64) float64
	// Name returns short, human-readable identifier of the Location
	Name() string

	neighbours() []Location
}

type basePosition struct {
	config    *graphConfig
	occupant  int // occupying vehicle's id
	request   chan request
	emergency chan request
}

func (pos *basePosition) getRequestChannel() chan<- request {
	return pos.request
}

func (pos *basePosition) getRWRequestChannel() chan request {
	return pos.request
}

type requestType int

const (
	_ requestType = iota
	take
	free
	reserve
	release
	repairStart
	repairDone
	check
)

//go:generate stringer -type requestType

type request struct {
	c        chan bool
	senderID int
	kind     requestType
}

type emergency struct {
	location Location
	handler  requestHandler
}

type handlerStatus struct {
	occupant      int
	position      Location
	failing       bool
	reservation   int
	repairStarted bool
	ctr           int
	handlers      map[requestType]func(*handlerStatus, request) bool
}

var defaultHandlers = map[requestType]func(*handlerStatus, request) bool{
	take:        doTake,
	free:        doFree,
	reserve:     doReserve,
	release:     doRelease,
	repairStart: doRepairStart,
	repairDone:  doRepairDone,
	check: func(s *handlerStatus, req request) bool {
		return !s.failing
	},
}

func (s handlerStatus) logf(format string, args ...interface{}) {
	tagFail := ""
	if s.failing {
		tagFail = " [Failing]"
	}
	prefix := fmt.Sprintf("[%s:%d]%s ", s.position.Name(), s.ctr, tagFail)
	log.Printf(prefix+format, args...)
}

// Handle handles position's communication with other network elements
func Handle(position Location, context *Graph) {
	var req request
	var response bool

	s := &handlerStatus{
		occupant:      -1,
		position:      position,
		failing:       false,
		reservation:   -1,
		repairStarted: false,
		ctr:           0,
		handlers:      defaultHandlers,
	}
	failures := make(chan bool)
	requests := position.getRWRequestChannel()
	go context.generateFailures(failures)
	for {
		select {
		case req = <-requests:
			s.ctr++
			s.logf("request: %v", req)
			response = s.handlers[req.kind](s, req)
			if req.kind == repairDone && response {
				// restart failure generator
				go context.generateFailures(failures)
			}
			req.c <- response
		case <-failures:
			s.failing = true
			s.logf("Sending emergency report")
			go func() { context.emergencyCtr <- report{delta: 1, key: position.Name()} }()
			go func() { context.Emergency <- emergency{position, position} }()
		}

	}
}

func (r request) String() string {
	return fmt.Sprintf("{%v, %d}", r.kind, r.senderID)
}
