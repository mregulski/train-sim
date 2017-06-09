package network

import (
	"fmt"
	"log"
	// "sync"
)

// Location is an element of the transport network that can be used by a vehicle
type Location interface {
	requestHandler
	// TravelTime returns time in hours required to get through this position at given speed
	TravelTime(speed float64) float64
	// Name returns short identifier of the Position
	Name() string
	// Handle is a supervisor method for the Position.
	// Handle()
	// GetRequestChannel returns a channel for sending requests to the Position
	// GetRequestChannel() chan<- request
	// getRWRequestChannel() chan request
	// Occupied returns true if the position is currently used by some vehicle
	Occupied() bool
	Occupant() int
	SetOccupant(int)
	neighbours() []Location
}

type basePosition struct {
	config    *GraphConfig
	occupant  int // occupying vehicle's id
	request   chan request
	emergency chan request
}

func (pos *basePosition) GetRequestChannel() chan<- request {
	return pos.request
}

func (pos *basePosition) GetRWRequestChannel() chan request {
	return pos.request
}

func (pos *basePosition) Occupant() int {
	return pos.occupant
}

func (pos *basePosition) SetOccupant(occ int) {
	pos.occupant = occ
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
	fail
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
}

// Handle handles position's communication with other network elements
func Handle(position Location, context *Graph) {
	var req request
	var response bool
	ctr := 0;
	// var lock = sync.Mutex{}
	// lock.Lock()
	s := handlerStatus{
		occupant:      -1,
		position:      position,
		failing:       false,
		reservation:   -1,
		repairStarted: false,
	}
	// lock.Unlock()
	failureStop := make(chan bool)
	requests := position.GetRWRequestChannel()
	// go context.generateFailures(requests, failureStop)
	for {
		req = <-requests
		ctr++
		if req.kind == fail { // does not expect a reply
			failureStop <- true // halt failures until this one's handled
			// lock.Lock()
			s.failing = true
			log.Printf("[%s:%d] Failure: sending emergency signal (status.failing: %v)",
				position.Name(), ctr, s.failing)
			go func() {context.emergencyCtr <- report{delta: 1, key: position.Name()}}()
			// lock.Unlock()
			go func() {context.Emergency <- emergency{position, position}}()
		} else {
			switch req.kind {
			case free:
				if s.failing {
					log.Printf("[%s:%d] Failing - cannot release vehicle", s.position.Name(), ctr)
					response = false
				} else {
					// CO TU SIĘ ODPIERDALA?!
					// stan wątku nie jest zsynchronizowany sam z sobą
					// log.Printf("[%s] occupant: %d, sender: %d", s.position.Name(), s.occupant, req.senderID)
					// if s.occupant == req.senderID { // no idea what the fuck happens here
					s.occupant = -1
					log.Printf("[%s:%d] Vehicle #%d left (occupant: %d)", s.position.Name(), ctr, req.senderID, s.occupant)
					response = true
				}
				// }
				// return false, s

			case take:
				if s.failing && s.reservation != req.senderID {
					log.Printf("[%s:%d] [Failing] refusing entry to Vehicle#%d", s.position.Name(), ctr, req.senderID)
					response = false
				} else if s.occupant == -1 || s.occupant == req.senderID {
					// !s.failing || reservation == req. senderID
					// no idea why 2nd part of the OR is necessary
					if s.reservation > 0 && s.reservation != req.senderID {
						log.Printf("[%s:%d] Location reserved by Vehicle#%d - refusing entry to #%d",
							s.position.Name(), s.reservation, req.senderID)
						response = false
					} else {
						log.Printf("[%s:%d] Vehicle #%d arrived", s.position.Name(), ctr, req.senderID)
						s.occupant = req.senderID
						response = true
					}
				} else {
					log.Printf("[%s:%d] Location already occupied by Vehicle#%d", s.position.Name(), ctr, s.occupant)
					response = false
				}

			case reserve:
				log.Printf("[%s:%d] Reservation request from %v", s.position.Name(), ctr, req.senderID)
				s.reservation = req.senderID
				log.Printf("[%s:%d] Reserving for %d", s.position.Name(), ctr, req.senderID)
				response = true

			case release:
				if s.reservation == req.senderID {
					log.Printf("[%s:%d] Releasing", s.position.Name(), ctr)
					s.reservation = -1
					response = true
				} else {
					response = false
				}

			case repairStart:
				if s.repairStarted {
					log.Printf("[%s:%d] [FAILING] Repair already started", s.position.Name(), ctr)
				} else {
					log.Printf("[%s:%d] [FAILING] Repair started", s.position.Name(), ctr)
				}
				s.repairStarted = true
				response = true

			case repairDone:
				// if status.repairStarted {
				log.Printf("[%s:%d] [FAILING] Repair finished", s.position.Name(), ctr)
				s.failing = false
				s.repairStarted = false
				log.Printf("[%s:%d] Back online", s.position.Name(), ctr)
				response = true
				go context.generateFailures(requests, failureStop)
				// } else {
				// 	log.Printf("[%s] [FAILING] Repair finished before starting?", status.position.Name())
				// 	req.c <- false
				// }
			case check:
				response = !s.failing
			}
			// END PASTE
			// // lock.Lock()
			// // log.Printf("[%s][DBG] status: %+v, req: %v", position.Name(), status, req.kind)
			// response, status = handleRequest(req, status)
			// // log.Printf("[%s][DBG] new status: %+v", position.Name(), status)
			// // log.Printf("[%s] request %v from #%d: %v", position.Name(), req.kind, req.senderID, response)
			// // log.Printf("%20s status: %+v", "", status)
			// lock.Unlock()
			req.c <- response
			// if req.kind == repairDone {
			// 	go context.generateFailures(requests, failureStop)
			// }
		}
	}
}

func handleRequest(req request, status handlerStatus) (bool, handlerStatus) {
	s := handlerStatus{
		occupant:      status.occupant,
		position:      status.position,
		failing:       status.failing,
		reservation:   status.reservation,
		repairStarted: status.repairStarted,
	}
	switch req.kind {
	case free:
		if s.failing {
			log.Printf("[%s] Failing - cannot release vehicle", s.position.Name())
			return false, s
		}
		// CO TU SIĘ ODPIERDALA?!
		// stan wątku nie jest zsynchronizowany sam z sobą
		// log.Printf("[%s] occupant: %d, sender: %d", s.position.Name(), s.occupant, req.senderID)
		// if s.occupant == req.senderID { // no idea what the fuck happens here
		log.Printf("[%s] Vehicle #%d left", s.position.Name(), req.senderID)
		s.occupant = -1
		return true, s
		// }
		// return false, s

	case take:
		if s.failing && s.reservation != req.senderID {
			log.Printf("[%s] [Failing] refusing entry to Vehicle#%d", s.position.Name(), req.senderID)
			return false, s
		}
		// !s.failing || reservation == req. senderID
		// no idea why 2nd part of the OR is necessary
		if s.occupant == -1 || s.occupant == req.senderID {
			if s.reservation > 0 && s.reservation != req.senderID {
				log.Printf("[%s] Location reserved by Vehicle#%d - refusing entry to #%d",
					s.position.Name(), s.reservation, req.senderID)
				return false, s
			}
			log.Printf("[%s] Vehicle #%d arrived", s.position.Name(), req.senderID)
			s.occupant = req.senderID
			return true, s
		}
		log.Printf("[%s] Location already occupied by Vehicle#%d", s.position.Name(), req.senderID)
		return false, s

	case reserve:
		log.Printf("[%s] Reservation request from %v", s.position.Name(), req.senderID)
		s.reservation = req.senderID
		log.Printf("[%s] Reserving for %d", s.position.Name(), req.senderID)
		return true, s

	case release:
		if s.reservation == req.senderID {
			log.Printf("[%s] Releasing", s.position.Name())
			s.reservation = -1
			return true, s
		}
		return false, s

	case repairStart:
		if s.repairStarted {
			log.Printf("[%s] [FAILING] Repair already started", s.position.Name())
		} else {
			log.Printf("[%s] [FAILING] Repair started", s.position.Name())
		}
		s.repairStarted = true
		return true, s

	case repairDone:
		// if status.repairStarted {
		log.Printf("[%s] [FAILING] Repair finished", s.position.Name())
		s.failing = false
		s.repairStarted = false
		log.Printf("[%s] Back online", s.position.Name())
		return true, s
		// } else {
		// 	log.Printf("[%s] [FAILING] Repair finished before starting?", status.position.Name())
		// 	req.c <- false
		// }
	case check:
		return !s.failing, s
	}
	panic("unhandled request")
}

func (r request) String() string {
	return fmt.Sprintf("{%v, %d}", r.kind, r.senderID)
}
