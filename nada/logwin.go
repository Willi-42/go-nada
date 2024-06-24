package nada

type packetEvent struct {
	lost         bool
	marked       bool
	queueBuildup bool
	size         uint64
	tsReceived   uint64
	lossRange    uint64 // for lost packets in a range
}

// LogWinQueue
type logWinQueue struct {
	elements   []packetEvent // all events of current window
	windowSize uint64        // size of the logging window in mirco seconds
	lossInts   lossInterval  // for the avg loss interval calculation

	numberPacketsSinceLoss uint64 // number of packets since the last loss occurred
	lastPn                 uint64

	// stats
	numberLostPackets   uint64
	numberMarkedPackets uint64
	numberPacketArrived uint64
	receivedBytes       uint64 // bytes received in current window
	queueBuildupCnt     uint64 // number of times a queue buildup was detected
}

// NewLogWinQueue creates a new logging window queue.
// Size has to be in micro seconds!
func NewLogWinQueue(sizeInMicroS uint64) *logWinQueue {
	return &logWinQueue{
		elements:   make([]packetEvent, 0),
		windowSize: sizeInMicroS,
		lossInts:   *newLossIntervall(8), // TODO: add to config
	}
}

func (q *logWinQueue) addPacketEvent(
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	q.elements = append(q.elements, packetEvent{
		lost:         false,
		marked:       marked,
		size:         size,
		tsReceived:   tsReceived,
		queueBuildup: queueBuildup,
	})
}

func (q *logWinQueue) addSkippedPN(pn, tsReceived uint64) {
	q.checkForGaps(pn, tsReceived)
	q.lastPn = pn
}

func (q *logWinQueue) NewMediaPacketRecieved(
	pn uint64,
	tsReceived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	if pn <= q.lastPn && pn != 0 {
		// packet arravied out of order
		// TODO: duplicated packet
		// TODO: should handle queue build up
		return
	}

	q.addPacketEvent(tsReceived, size, marked, queueBuildup)
	q.receivedBytes += size
	q.numberPacketArrived++
	q.numberPacketsSinceLoss++
	q.lossInts.addPacket()

	if marked {
		q.numberMarkedPackets++
	}

	if queueBuildup {
		q.queueBuildupCnt++
	}

	q.checkForGaps(pn, tsReceived)

	q.lastPn = pn
}

func (q *logWinQueue) checkForGaps(pn, tsReceived uint64) {
	// skip gap calc for first packet
	if pn == 0 {
		return
	}

	// missing packets are considered lost
	gapSize := pn - q.lastPn - 1

	// packet gap
	if gapSize != 0 {
		q.addLossEvent(tsReceived, gapSize)
		q.numberLostPackets += gapSize
		q.numberPacketsSinceLoss = 1
		q.lossInts.addLoss(gapSize)
	}
}

func (q *logWinQueue) addLossEvent(ts, lossRange uint64) {
	q.elements = append(q.elements, packetEvent{
		lost:       true,
		tsReceived: ts,
		lossRange:  lossRange,
		marked:     false,
	})
}

func (q *logWinQueue) updateStats(currentTime uint64) {
	if currentTime <= q.windowSize {
		return
	}

	threashold := currentTime - q.windowSize
	discardIndes := 0

	for _, event := range q.elements {
		if event.tsReceived <= threashold {
			discardIndes++

			if event.lost {
				q.numberLostPackets -= event.lossRange
				continue
			}

			if event.marked {
				q.numberMarkedPackets--
			}

			if event.queueBuildup {
				q.queueBuildupCnt--
			}

			q.numberPacketArrived--
			q.receivedBytes -= event.size
		}
	}

	// drop old elements
	q.elements = q.elements[discardIndes:] // TODO: very inefficient
}
