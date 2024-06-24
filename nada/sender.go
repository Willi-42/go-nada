package nada

import "time"

type Sender struct {
	refernceRate uint64 // Reference rate based on network congestion
	xPerv        uint64 // Previous value of aggregate congestion signal
	lastReport   uint64 // in micro sec

	config *Config
}

func NewSender(config Config) Sender {
	configPopulated := populateConfig(&config)

	return Sender{
		refernceRate: configPopulated.MinRate,
		xPerv:        0,
		config:       configPopulated,
	}
}

// FeedbackReport calculates the new rate with the feedback from the receiver.
// xCurr, recvRate, rampUpMode are from the reciver feedback.
// rtt is the current rtt in micro seconds.
func (s *Sender) FeedbackReport(xCurr uint64, recvRate uint64, rampUpMode bool, rtt uint64) uint64 {
	currTime := uint64(time.Now().UnixMicro())

	// default feedback interval
	delta := s.config.FeedbackDelta

	// skip for first feeback
	if s.lastReport != 0 {
		delta = currTime - s.lastReport
	}

	if rampUpMode {
		s.refernceRate = rampUpRate(*s.config, rtt, s.refernceRate, recvRate)
	} else {
		s.refernceRate = gradualUpdateRate(*s.config, s.refernceRate, xCurr, s.xPerv, delta)
	}

	// TODO: clip rate r_ref within the range of minimum rate (RMIN) and maximum rate (RMAX).

	s.xPerv = xCurr
	s.lastReport = currTime

	return s.refernceRate
}
