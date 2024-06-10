package nada

type packetEvent struct {
	lost      bool
	marked    bool
	size      uint64
	tsRecived uint64
	lossRange uint64 // for lost packets in a range
}

// LogWinQueue
type LogWinQueue struct {
	elements     []packetEvent
	sizeInMicroS uint64

	numberOfPacketsSinceLastLoss uint64
	lastPn                       uint64

	// stats
	numberLostPackets   uint64
	numberMarkedPackets uint64
	numberPacketArrived uint64
	accumulatedSize     uint64
}

// NewLogWinQueue creates a new logging window queue.
// Size has to be in micro seconds!
func NewLogWinQueue(sizeInMicroS uint64) *LogWinQueue {
	return &LogWinQueue{
		elements:     make([]packetEvent, 0),
		sizeInMicroS: sizeInMicroS,
	}
}

func (q *LogWinQueue) addLostEvent(ts, lossRange uint64) {
	q.elements = append(q.elements, packetEvent{lost: true, tsRecived: ts, lossRange: lossRange})
}

func (q *LogWinQueue) addPacketEvent(pn uint64, tsRecived uint64, size uint64, marked bool) {
	q.elements = append(q.elements, packetEvent{
		lost:      false,
		marked:    marked,
		size:      size,
		tsRecived: tsRecived,
	})
}

func (q *LogWinQueue) NewMediaPacketRecieved(pn uint64, tsRecived uint64, size uint64, marked bool) {
	if pn <= q.lastPn && pn != 0 {
		// packet arravied out of order
		// TODO: duplicated packet
		return
	}

	q.addPacketEvent(pn, tsRecived, size, marked)
	q.accumulatedSize += size
	q.numberPacketArrived++
	q.numberOfPacketsSinceLastLoss++

	if marked {
		q.numberMarkedPackets++
	}

	// skip gap calc for first packet
	if pn == 0 {
		return
	}

	// missing packets are considered lost
	pnGap := pn - q.lastPn - 1

	if pnGap != 0 {
		q.addLostEvent(tsRecived, pnGap)
		q.numberLostPackets += pnGap
		q.numberOfPacketsSinceLastLoss = 1
	}

	q.lastPn = pn
}

func (q *LogWinQueue) updateStats(currentTime uint64) {
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

			q.numberPacketArrived--
			q.accumulatedSize -= event.size
		}
	}

	// drop old elements
	q.elements = q.elements[discardIndes:] // TODO: very inefficient
}
