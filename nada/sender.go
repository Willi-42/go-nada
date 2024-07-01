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
		prevRate: configPopulated.MinRate,
		config:   configPopulated,
	}
}

// FeedbackReport calculates the new rate with the feedback from the receiver.
// xCurr, recvRate, rampUpMode are from the receiver feedback.
// rtt is the current rtt in micro seconds.
func (s *Sender) FeedbackReport(xCurr uint64, recvRate uint64, rampUpMode bool, rtt uint64) (newRate uint64) {
	currTime := uint64(time.Now().UnixMicro())

	// default feedback interval
	delta := s.config.FeedbackDelta

	// skip for first feeback
	if s.lastReport != 0 {
		delta = currTime - s.lastReport
	}

	if rampUpMode {
		newRate = rampUpRate(*s.config, rtt, s.prevRate, recvRate)
	} else {
		newRate = gradualUpdateRate(*s.config, s.prevRate, xCurr, s.xPerv, delta)
	}

	// clip rate within minimum rate and maximum rate
	newRate = max(s.config.MinRate, newRate)
	newRate = min(s.config.MaxRate, newRate)

	s.prevRate = newRate
	s.xPerv = xCurr
	s.lastReport = currTime

	return newRate
}
