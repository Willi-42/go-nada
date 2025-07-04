package nada

import (
	"sync"
	"time"

	"github.com/Willi-42/go-nada/nada/windows"
)

type SenderSLD struct {
	prevRate   uint64 // Previous reference rate based on network congestion
	xPerv      uint64 // Previous value of aggregate congestion signal
	lastReport uint64 // in micro sec

	lossRatio    float64 // estimated packet loss ratio (p_loss)
	markingRatio float64 // estimated packet ECN marking ratio (p_mark)

	mtx sync.Mutex

	config *Config
	logWin *windows.LogWindow
}

func NewSenderSLD(config Config) SenderSLD {
	configPopulated := populateConfig(&config)
	logWinSize := configPopulated.LogWin

	return SenderSLD{
		prevRate: configPopulated.StartRate,
		config:   configPopulated,
		logWin:   windows.NewLogWindow(logWinSize, 8), // TODO: add to config
	}
}

// PacketDelivered register a delivered packet.
// QueueBuildUp is calculated at receiver.
// Use arrival ts of Ack for the LogWindow.
func (s *SenderSLD) PacketDelivered(
	packetNumber uint64,
	ackTs uint64,
	packetSize uint64,
	marked bool,
) {
	s.mtx.Lock()
	s.logWin.NewMediaPacketRecievedNoGapCheck(packetNumber, ackTs, packetSize, marked, false)

	s.logWin.UpdateStats(ackTs)

	s.mtx.Unlock()
}

// LostPacket registers a lost packet
func (r *SenderSLD) LostPacket(pn, tsReceived uint64) {
	r.mtx.Lock()
	r.logWin.AddLostPacket(pn, tsReceived)
	r.mtx.Unlock()
}

// FeedbackReport calculates the new rate with the feedback from the receiver.
// recvRate, delay, queueBuildup are from the receiver feedback.
// rtt is the current rtt in micro seconds.
func (s *SenderSLD) FeedbackReport(feedback FeedbackSLD, rtt uint64) (newRate, xCurr uint64) { // TODO: remove xcurr from return
	s.mtx.Lock()
	currTime := uint64(time.Now().UnixMicro())

	// default feedback interval
	delta := s.config.FeedbackDelta

	// skip for first feedback
	if s.lastReport != 0 {
		delta = currTime - s.lastReport
	}

	// update congestion values
	updatedDelay := wrapQDelay(*s.config, feedback.Delay, s.logWin)

	lostPackets := s.logWin.LostPackets()
	markedpackets := s.logWin.MarkedPackets()

	totoalPackets := s.logWin.ArrivedPackets() + lostPackets

	s.mtx.Unlock()

	s.lossRatio = smoothedRatio(*s.config, lostPackets, totoalPackets, s.lossRatio)
	s.markingRatio = smoothedRatio(*s.config, markedpackets, totoalPackets, s.markingRatio)

	xCurr = aggregateCng(*s.config, updatedDelay, s.markingRatio, s.lossRatio)

	// calc new rate
	if lostPackets == 0 && !feedback.QueueBuildup {
		// RampUp
		newRate = rampUpRate(*s.config, rtt, s.prevRate, feedback.RecvRate)
	} else {
		newRate = gradualUpdateRate(*s.config, s.prevRate, xCurr, s.xPerv, delta)
	}

	// clip rate within minimum rate and maximum rate
	newRate = max(s.config.MinRate, newRate)
	newRate = min(s.config.MaxRate, newRate)

	s.prevRate = newRate
	s.xPerv = xCurr
	s.lastReport = currTime

	return newRate, xCurr
}
