package nada

import "time"

type Sender struct {
	refernceRate uint64 // Reference rate based on network congestion
	rtt          uint64 // Estimated round-trip time
	xPerv        uint64 // Previous value of aggregate congestion signal
	lastReport   uint64

	config *Config
}

func NewSender(config Config) Sender {
	configPopulated := populateConfig(&config)
	// logWinSize := configPopulated.LogWin * 1000 // convert to micro sec

	return Sender{
		refernceRate: configPopulated.MinRate,
		rtt:          0,
		xPerv:        0,
		config:       configPopulated,
	}
}

func (s *Sender) FeedbackReport(xCurr uint64, recvRate uint64, rampUpMode bool) {
	currTime := uint64(time.Now().UnixMicro())

	// TODO: update rtt

	// measure feedback interval
	delta := currTime - s.lastReport

	if rampUpMode {
		s.refernceRate = rampUpRate(*s.config, s.rtt, s.refernceRate, recvRate)
	} else {
		s.refernceRate = gradualUpdateRate(*s.config, s.refernceRate, xCurr, s.xPerv, delta)
	}

	// TODO: clip rate r_ref within the range of minimum rate (RMIN) and maximum rate (RMAX).

	s.xPerv = xCurr
	s.lastReport = currTime
}
