package windows

type packetEvent struct {
	lost         bool
	marked       bool
	queueBuildup bool
	size         uint64
	tsReceived   uint64
	lossRange    uint64 // for lost packets in a range
	pn           uint64
}

// LogWindow is a logging window
type LogWindow struct {
	elements   []packetEvent // all events of current window
	windowSize uint64        // size of the logging window in mirco seconds
	lossInts   *lossInterval // for the avg loss interval calculation

	lastPn uint64

	// stats of current window
	arrivedPackets  uint64
	markedPackets   uint64
	lostPackets     uint64
	receivedBits    uint64
	queueBuildupCnt uint64 // number of times a queue buildup was detected

	gotFirstPacket bool // to register the sequence number of our first packet as the starting point
}

// NewLogWindow creates a new logging window queue.
// Size has to be in micro seconds!
func NewLogWindow(sizeInMicroS uint64, lossIntervalSize int) *LogWindow {
	return &LogWindow{
		elements:       make([]packetEvent, 0),
		windowSize:     sizeInMicroS,
		lossInts:       newLossIntervall(lossIntervalSize),
		gotFirstPacket: false,
	}
}

func (q *LogWindow) LostPackets() uint64 {
	return q.lostPackets
}

func (q *LogWindow) MarkedPackets() uint64 {
	return q.markedPackets
}

func (q *LogWindow) ArrivedPackets() uint64 {
	return q.arrivedPackets
}

func (q *LogWindow) ReceivedBits() uint64 {
	return q.receivedBits
}

func (q *LogWindow) PacketsSinceLoss() (uint64, bool /* got loss*/) {
	return q.lossInts.currentInt()
}

// QueueBuildup in current window?
func (q *LogWindow) QueueBuildup() bool {
	return q.queueBuildupCnt != 0
}

func (q *LogWindow) AvgLossInterval() float64 {
	return q.lossInts.avgLossInt()
}

func (q *LogWindow) AddEmptyPacket(pn, tsReceived uint64) {
	if !q.gotFirstPacket {
		q.lastPn = pn
		q.gotFirstPacket = true

	} else if pn <= q.lastPn {
		// packet arravied out of order
		// TODO: duplicated packet
		// TODO: should handle queue build up
		return
	}

	q.checkForGaps(pn, tsReceived)
	q.lastPn = pn
}

// AddLostPacket to register a loss if detected by app.
// For loss detection at sender side
func (q *LogWindow) AddLostPacket(pn, tsReceived uint64) {
	q.addLossEvent(tsReceived, 1, pn, true)
	q.lastPn = pn
}

func (q *LogWindow) NewMediaPacketRecieved(
	pn uint64,
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	if !q.gotFirstPacket {
		q.lastPn = pn
		q.gotFirstPacket = true

	} else if pn <= q.lastPn {
		// packet arravied out of order
		// TODO: duplicated packet
		// TODO: should handle queue build up
		return
	}

	q.checkForGaps(pn, tsReceived)

	q.NewMediaPacketRecievedNoGapCheck(pn, tsReceived, size, marked, queueBuildup)
}

func (q *LogWindow) NewMediaPacketRecievedNoGapCheck(
	pn uint64,
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	q.addPacketEvent(tsReceived, size, marked, queueBuildup, pn)
	q.receivedBits += size
	q.arrivedPackets++

	if marked {
		q.markedPackets++
	}

	if queueBuildup {
		q.queueBuildupCnt++
	}

	q.lastPn = pn
}

func (q *LogWindow) UpdateStats(currentTime uint64) {
	if currentTime <= q.windowSize {
		return
	}

	threashold := currentTime - q.windowSize
	discardIndex := 0

	for _, event := range q.elements {
		if event.tsReceived <= threashold {
			discardIndex++

			if event.lost {
				q.lostPackets -= event.lossRange
				continue
			}

			if event.marked {
				q.markedPackets--
			}

			if event.queueBuildup {
				q.queueBuildupCnt--
			}

			q.arrivedPackets--
			q.receivedBits -= event.size
		}
	}

	// drop old elements
	q.elements = q.elements[discardIndex:] // TODO: maybe inefficient
}

func (q *LogWindow) checkForGaps(pn, tsReceived uint64) {
	// skip gap calc for first packet and duplicated packets
	if pn == q.lastPn {
		return
	}

	// missing packets are considered lost
	gapSize := pn - q.lastPn - 1

	// packet gap
	if gapSize != 0 {
		q.addLossEvent(tsReceived, gapSize, pn, false)
	}
}

func (q *LogWindow) addLossEvent(ts, gapSize, pn uint64, sortedInsert bool) {
	lossEvent := packetEvent{
		lost:       true,
		tsReceived: ts,
		lossRange:  gapSize,
		marked:     false,
		pn:         pn,
	}

	if !sortedInsert {
		q.elements = append(q.elements, lossEvent)
	} else {
		// sored instert based on tsReceived
		inserted := false
		for i, event := range q.elements {
			if event.tsReceived >= lossEvent.tsReceived {
				// Insert before event at index i
				q.elements = append(q.elements[:i], append([]packetEvent{lossEvent}, q.elements[i:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			// If not inserted, append at the end
			q.elements = append(q.elements, lossEvent)
		}
	}

	q.lostPackets += gapSize
	q.lossInts.addLoss(gapSize)
}

func (q *LogWindow) addPacketEvent(tsReceived uint64, size uint64, marked bool, queueBuildup bool, pn uint64) {
	q.elements = append(q.elements, packetEvent{
		lost:         false,
		marked:       marked,
		size:         size,
		tsReceived:   tsReceived,
		queueBuildup: queueBuildup,
		pn:           pn,
	})

	q.lossInts.addPacket()
}
