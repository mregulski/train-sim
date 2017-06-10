package network

func doFree(s *handlerStatus, req request) bool {
	if s.failing {
		s.logf("Cannot release vehicle #%d", req.senderID)
		return false
	}
	if s.occupant == req.senderID {
		s.occupant = -1
		s.logf("Vehicle #%d left", req.senderID)
		return true
	}
	s.logf("Vehicle #%d wants to leave but occupant is #%d", req.senderID, s.occupant)
	return false

}

func doTake(s *handlerStatus, req request) bool {
	if s.failing && s.reservation != req.senderID {
		s.logf("refusing entry to vehicle #%d", req.senderID)
		return false
	}
	if s.occupant == -1 || s.occupant == req.senderID {
		if s.reservation > 0 && s.reservation != req.senderID {
			s.logf("Refusing entry to vehicle #%d - location reserved by Vehicle#%d", req.senderID, s.reservation)
			return false
		}
		s.logf("Vehicle #%d arrived", req.senderID)
		s.occupant = req.senderID
		return true

	}
	s.logf("Refusing entry to vehicle #%d - location already occupied by Vehicle#%d", req.senderID, s.occupant)
	return false
}

func doRelease(s *handlerStatus, req request) bool {
	if s.reservation == req.senderID {
		s.logf("Releasing reservation")
		s.reservation = -1
		return true
	}
	s.logf("Rejecting 'release' from vehicle #%d - reserved by vehicle #%d", req.senderID, s.reservation)
	return false
}

func doReserve(s *handlerStatus, req request) bool {
	if s.reservation == -1 {
		s.logf("Reserving for %d", req.senderID)
		s.reservation = req.senderID
		return true
	}
	s.logf("Rejecting reservation by vehicle #%d - already reserved by vehicle #%d", req.senderID, s.reservation)
	return false
}

func doRepairStart(s *handlerStatus, req request) bool {
	if s.repairStarted {
		s.logf("Repair has alredy been started")
		return false
	}
	s.logf("Repair started")
	s.repairStarted = true
	return true
}

func doRepairDone(s *handlerStatus, req request) bool {
	if s.repairStarted {
		s.logf("Repair finished")
		s.failing = false
		s.repairStarted = false
		s.logf("Back online")
		return true
	}
	s.logf("Ignoring repairDone: repairStart is required first")
	return false
}
