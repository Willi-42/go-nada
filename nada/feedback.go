package nada

import (
	"fmt"
	"time"
)

type FeedbackRLD struct {
	RecvRate   uint64
	XCurr      uint64
	RampUpMode bool
}

type FeedbackSLD struct {
	RecvRate     uint64
	Delay        uint64
	QueueBuildup bool
}

func (f FeedbackRLD) String() string {
	return fmt.Sprintf("%v;%v;%v", f.RecvRate, f.XCurr, f.RampUpMode)
}

func (f FeedbackSLD) String() string {
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
}
