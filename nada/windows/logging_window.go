package windows

type packetEvent struct {
	lost         bool
	marked       bool
	queueBuildup bool
	size         uint64
	tsReceived   uint64
	lossRange    uint64 // for lost packets in a range
	pNr          uint64
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

func (l *LogWindow) LostPackets() uint64 {
	return l.lostPackets
}

func (l *LogWindow) MarkedPackets() uint64 {
	return l.markedPackets
}

func (l *LogWindow) ArrivedPackets() uint64 {
	return l.arrivedPackets
}

func (l *LogWindow) ReceivedBits() uint64 {
	return l.receivedBits
}

func (l *LogWindow) PacketsSinceLoss() (uint64, bool /* got loss*/) {
	return l.lossInts.currentInt()
}

// QueueBuildup in current window?
func (l *LogWindow) QueueBuildup() bool {
	return l.queueBuildupCnt != 0
}

func (l *LogWindow) AvgLossInterval() float64 {
	return l.lossInts.avgLossInt()
}

func (l *LogWindow) AddEmptyPacket(pNr, tsReceived uint64) {
	if !l.gotFirstPacket {
		l.lastPn = pNr
		l.gotFirstPacket = true

	} else if pNr > l.lastPn {
		// only for packets that arrived in order
		l.checkForGaps(pNr, tsReceived)
		l.lastPn = pNr
	}
}

// AddLostPacket to register a loss if detected by app.
// For loss detection at sender side
func (l *LogWindow) AddLostPacket(pNr, tsReceived uint64) {
	l.addLossEvent(pNr, tsReceived, 1, true)
	l.lastPn = pNr
}

func (l *LogWindow) NewMediaPacketRecieved(
	pNr uint64,
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	if !l.gotFirstPacket {
		l.lastPn = pNr
		l.gotFirstPacket = true

	} else if pNr <= l.lastPn {
		// do not register packets that arrived out of order
		return
	}

	l.checkForGaps(pNr, tsReceived)
	l.NewMediaPacketRecievedNoGapCheck(pNr, tsReceived, size, marked, queueBuildup)
}

func (l *LogWindow) NewMediaPacketRecievedNoGapCheck(
	pNr uint64,
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	ok := l.addPacketEvent(pNr, tsReceived, size, marked, queueBuildup)
	if !ok {
		// packet duplicated; ignore
		return
	}

	l.receivedBits += size
	l.arrivedPackets++

	if marked {
		l.markedPackets++
	}

	if queueBuildup {
		l.queueBuildupCnt++
	}

	if pNr < l.lastPn {
		// packet arrived out of order, do not update lastPn
		return
	}

	l.lastPn = pNr
}

func (l *LogWindow) UpdateStats(currentTime uint64) {
	if currentTime <= l.windowSize {
		return
	}

	threashold := currentTime - l.windowSize
	discardIndex := 0

	for _, event := range l.elements {
		if event.tsReceived <= threashold {
			discardIndex++

			if event.lost {
				l.lostPackets -= event.lossRange
				continue
			}

			if event.marked {
				l.markedPackets--
			}

			if event.queueBuildup {
				l.queueBuildupCnt--
			}

			l.arrivedPackets--
			l.receivedBits -= event.size
		}
	}

	// drop old elements
	l.elements = l.elements[discardIndex:] // TODO: maybe inefficient
}

func (l *LogWindow) checkForGaps(pNr, tsReceived uint64) {
	// skip gap calc for first packet and duplicated packets
	if pNr == l.lastPn {
		return
	}

	// missing packets are considered lost
	gapSize := pNr - l.lastPn - 1

	// packet gap
	if gapSize != 0 {
		pnBeforeLoss := uint64(0)
		if pNr > 0 {
			pnBeforeLoss = pNr - 1
		}
		l.addLossEvent(pnBeforeLoss, tsReceived, gapSize, false)
	}
}

func (l *LogWindow) addLossEvent(lastPnInLossRange, ts, gapSize uint64, sortedInsert bool) {
	lossEvent := packetEvent{
		lost:       true,
		tsReceived: ts,
		lossRange:  gapSize,
		marked:     false,
		pNr:        lastPnInLossRange,
	}

	if !sortedInsert {
		l.elements = append(l.elements, lossEvent)
	} else {
		ok := l.sortedInsertPacketEvent(lossEvent)
		if !ok {
			// duplicated loss event; nothing todo
			return
		}
	}

	l.lostPackets += gapSize
	l.lossInts.addLoss(gapSize)
}

func (l *LogWindow) addPacketEvent(pNr, tsReceived, size uint64, marked bool, queueBuildup bool) (ok bool) {
	ok = l.sortedInsertPacketEvent(packetEvent{
		lost:         false,
		marked:       marked,
		size:         size,
		tsReceived:   tsReceived,
		queueBuildup: queueBuildup,
		pNr:          pNr,
	})
	if !ok {
		return false
	}

	l.lossInts.addPacket()

	return true
}

// sortedInsertPacketEvent inserts a packet event in sorted order based on tsReceived.
// Does not insert duplicated packets if they have the same received timestamp.
func (l *LogWindow) sortedInsertPacketEvent(newEvent packetEvent) (ok bool) {
	inserted := false
	for i, currEvent := range l.elements {
		if currEvent.tsReceived == newEvent.tsReceived && currEvent.pNr == newEvent.pNr {
			// packet duplicated; do not instert
			return false
		}

		if currEvent.tsReceived > newEvent.tsReceived {
			// Insert before first event that happend after newEvent
			l.elements = append(l.elements[:i], append([]packetEvent{newEvent}, l.elements[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		// If no later event found, append at the end
		l.elements = append(l.elements, newEvent)
	}

	return true
}
