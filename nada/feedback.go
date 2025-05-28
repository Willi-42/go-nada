package nada

import "fmt"

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
