package nada

import (
	"fmt"
	"time"
)

type Feedback struct {
	RecvRate     uint64
	Delay        uint64
	QueueBuildup bool
}

func (f Feedback) String() string {
	return fmt.Sprintf("%v;%v;%v", f.RecvRate, f.Delay, f.QueueBuildup)
}

// Acknowledgment contains arrival information about one packet.
// Sender-only NADA requires these Acknowledgments.
type Acknowledgment struct {
	SeqNr     uint64
	Departure time.Time
	Arrival   time.Time
	SizeBit   uint64
	Marked    bool
	Arrived   bool
}
