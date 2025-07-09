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

// Feedback for sender only NADA
type FeedbackSO struct {
	PacketNumber  uint64
	sentTs        time.Time
	RecvTs        time.Time
	PacketSizeBit uint64
	Marked        bool
}
