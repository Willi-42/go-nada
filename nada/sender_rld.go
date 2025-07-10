package nada

import (
	"time"
)

type Sender struct {
	prevRate   uint64 // Previous reference rate based on network congestion
	xPerv      uint64 // Previous value of aggregate congestion signal
	lastReport uint64 // in micro sec

	config *Config
}

func NewSender(config Config) Sender {
	configPopulated := populateConfig(&config)

	return Sender{
		prevRate: configPopulated.StartRate,
		config:   configPopulated,
	}
}

// FeedbackReport calculates the new rate with the feedback from the receiver.
// xCurr, recvRate, rampUpMode are from the receiver feedback.
// rtt is the current rtt in micro seconds.
func (s *Sender) FeedbackReport(feedback FeedbackRLD, rtt time.Duration) (newRate uint64) {
	currTime := uint64(time.Now().UnixMicro())

	// default feedback interval
	delta := s.config.FeedbackDelta

	// skip for first feedback
	if s.lastReport != 0 {
		delta = currTime - s.lastReport
	}

	if feedback.RampUpMode {
		newRate = rampUpRate(*s.config, uint64(rtt.Microseconds()), s.prevRate, feedback.RecvRate)
	} else {
		newRate = gradualUpdateRate(*s.config, s.prevRate, feedback.XCurr, s.xPerv, delta)
	}

	// clip rate within minimum rate and maximum rate
	newRate = max(s.config.MinRate, newRate)
	newRate = min(s.config.MaxRate, newRate)

	s.prevRate = newRate
	s.xPerv = feedback.XCurr
	s.lastReport = currTime

	return newRate
}
