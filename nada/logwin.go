package nada

type packetEvent struct {
	lost         bool
	marked       bool
	queueBuildup bool
	size         uint64
	tsRecived    uint64
	lossRange    uint64 // for lost packets in a range
}

// LogWinQueue
type logWinQueue struct {
	elements     []packetEvent
	sizeInMicroS uint64
	lossInt      lossInterval

	numberPacketsSinceLoss uint64 // number of packets since the last loss occurred
	lastPn                 uint64

	// stats
	numberLostPackets   uint64
	numberMarkedPackets uint64
	numberPacketArrived uint64
	totalSize           uint64 // bytes recived in current window
	queueBuildupCnt     uint64 // number of times a queue buildup was detected
}

// NewLogWinQueue creates a new logging window queue.
// Size has to be in micro seconds!
func NewLogWinQueue(sizeInMicroS uint64) *logWinQueue {
	return &logWinQueue{
		elements:     make([]packetEvent, 0),
		sizeInMicroS: sizeInMicroS,
		lossInt:      *newLossIntervall(8), // TODO: add to config
	}
}

func (q *logWinQueue) addPacketEvent(
	tsRecived uint64,
	size uint64,
	marked bool,
	queueBuildup bool,
) {
	q.elements = append(q.elements, packetEvent{
		lost:         false,
		marked:       marked,
		size:         size,
		tsRecived:    tsRecived,
		queueBuildup: queueBuildup,
	})
}

func (q *logWinQueue) addSkippedPN(pn, tsRecived uint64) {
	q.addLostGap(pn, tsRecived)
	q.lastPn = pn
}

func (q *logWinQueue) NewMediaPacketRecieved(
	pn uint64,
	tsRecived uint64,
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

	q.addPacketEvent(tsRecived, size, marked, queueBuildup)
	q.totalSize += size
	q.numberPacketArrived++
	q.numberPacketsSinceLoss++
	q.lossInt.addPacket()

	if marked {
		q.numberMarkedPackets++
	}

	if queueBuildup {
		q.queueBuildupCnt++
	}

	q.addLostGap(pn, tsRecived)

	q.lastPn = pn
}

func (q *logWinQueue) addLostGap(pn, tsRecived uint64) {
	// skip gap calc for first packet
	if pn == 0 {
		return
	}

	// missing packets are considered lost
	gapSize := pn - q.lastPn - 1

	// packet gap
	if gapSize != 0 {
		q.addLostEvent(tsRecived, gapSize)
		q.numberLostPackets += gapSize
		q.numberPacketsSinceLoss = 1
		q.lossInt.addLoss(gapSize)
	}
}

func (q *logWinQueue) addLostEvent(ts, lossRange uint64) {
	q.elements = append(q.elements, packetEvent{
		lost:      true,
		tsRecived: ts,
		lossRange: lossRange,
		marked:    false,
	})
}

func (q *logWinQueue) updateStats(currentTime uint64) {
	if currentTime <= q.sizeInMicroS {
		return
	}

	threashold := currentTime - q.sizeInMicroS
	discardIndes := 0

	for _, event := range q.elements {
		if event.tsRecived <= threashold {
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
			q.totalSize -= event.size
		}
	}

	// drop old elements
	q.elements = q.elements[discardIndes:] // TODO: very inefficient
}
