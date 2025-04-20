package nada

import (
	"math"

	"github.com/Willi-42/go-nada/nada/windows"
)

// Receiver: all timestamps are in microseconds
type Receiver struct {
	baseDelay uint64 // estimated baseline delay (d_base)
	qDelay    uint64 // estimated queuing delay (d_queue)
	recvRate  uint64 // receiving rate (r_recv)

	lossRatio    float64 // estimated packet loss ratio (p_loss)
	markingRatio float64 // estimated packet ECN marking ratio (p_mark)

	config   *Config
	logWin   *windows.LogWindow
	delayWin *windows.DelayWindow // used to filter delay samples
}

func NewReceiver(config Config) Receiver {
	configPopulated := populateConfig(&config)
	logWinSize := configPopulated.LogWin

	return Receiver{
		baseDelay: math.MaxUint64, // set to infinity
		config:    configPopulated,
		logWin:    windows.NewLogWindow(logWinSize, 8), // TODO: add to config
		delayWin:  windows.NewDelayWindow(15),          // TODO: add to config
	}
}

// PacketArrivedWithoutTs can be used to register the
// arrival of packets without a ts, e.g. ack only packets.
// Have to be registered, otherwise considered lost.
func (r *Receiver) PacketArrivedWithoutTs(packetNumber, recvTs uint64) {
	r.logWin.AddEmptyPacket(packetNumber, recvTs)
}

// PacketArrived registers a new arrived packet.
// sentTs and recvTs are timestamps in microseconds.
// packetSize is the size of the packet in bits.
// marked: packet got ECN
func (r *Receiver) PacketArrived(
	packetNumber uint64,
	sentTs uint64,
	recvTs uint64,
	packetSize uint64,
	marked bool,
) {
	// current one-way delay (d_fwd)
	oneWayDelay := recvTs - sentTs

	// update base delay
	r.baseDelay = min(r.baseDelay, oneWayDelay)

	// TODO: recompute base delay from time to time e.g. 10min

	// update queue delay
	currDelay := oneWayDelay - r.baseDelay

	// filter qdelay with min filter
	// compare: https://www.rfc-editor.org/rfc/rfc8698.html#name-method-for-delay-loss-and-m
	r.delayWin.AddSample(currDelay)
	r.qDelay = r.delayWin.MinDelay()

	// check for queue build-up
	queueBuildup := false
	if r.qDelay >= r.config.QEPS {
		queueBuildup = true
	}

	// update logwin
	r.logWin.NewMediaPacketRecieved(packetNumber, recvTs, packetSize, marked, queueBuildup)
	r.logWin.UpdateStats(recvTs)

	// calculate loss/marking ratio
	totoalPackets := r.logWin.ArrivedPackets() + r.logWin.LostPackets()
	r.lossRatio = smoothedRatio(*r.config, r.logWin.LostPackets(), totoalPackets, r.lossRatio)
	r.markingRatio = smoothedRatio(*r.config, r.logWin.MarkedPackets(), totoalPackets, r.markingRatio)

	// update reciving rate
	// TODO: this might overflow
	recvBitsSeconds := r.logWin.ReceivedBits() * 1000000 // to convert micro seconds to seconds

	r.recvRate = recvBitsSeconds / r.config.LogWin
}

// GenerateFeedbackRLD: On time to send a new feedback report (t_curr - t_last > DELTA)
// Returns reciving rate, aggregated congestion signal and rampUpMode.
func (r *Receiver) GenerateFeedbackRLD() (recvRate uint64, xCurr uint64, rampUpMode bool) {
	recvRate = r.recvRate

	wrappedDelay := wrapQDelay(*r.config, r.qDelay, r.logWin)

	// calculate aggregate congestion signal x_curr
	xCurr = aggregateCng(*r.config, wrappedDelay, r.markingRatio, r.lossRatio)

	// determine mode of rate adaptation for sender: rmode
	// if no packet loss in logwin and no queue build up in current LOGWIN
	if r.logWin.LostPackets() == 0 && !r.logWin.QueueBuildup() {
		rampUpMode = true
	}

	return
}

// GenerateFeedbackSLD for loss detection at sender
func (r *Receiver) GenerateFeedbackSLD() (recvRate uint64, delay uint64, queueBuildup bool) {
	recvRate = r.recvRate
	delay = r.qDelay
	queueBuildup = r.logWin.QueueBuildup()

	return recvRate, delay, queueBuildup
}
