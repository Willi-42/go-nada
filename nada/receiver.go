package nada

import (
	"math"
	"time"

	"github.com/Willi-42/go-nada/nada/windows"
)

// Receiver: all timestamps are in microseconds
type Receiver struct {
	baseDelay    uint64 // estimated baseline delay (d_base)
	qDelay       uint64 // estimated queuing delay (d_queue)
	recvRate     uint64 // receiving rate (r_recv)
	lastArraival uint64 // timestamp of last arrived packet

	config   *Config
	logWin   *windows.LogWindow
	delayWin *windows.DelayWindow // used to filter delay samples
}

func NewReceiver(config Config) Receiver {
	configPopulated := populateConfig(&config)
	logWinSize := configPopulated.LogWin

	return Receiver{
		baseDelay:    math.MaxUint64, // set to infinity
		config:       configPopulated,
		logWin:       windows.NewLogWindow(logWinSize, 8), // TODO: add to config
		delayWin:     windows.NewDelayWindow(15),          // TODO: add to config
		lastArraival: 0,
	}
}

// PacketArrived registers a new arrived packet.
// sentTs and recvTs are timestamps in microseconds.
// packetSize is the size of the packet in bits.
// marked: packet got ECN
func (r *Receiver) PacketArrived(
	packetNumber uint64,
	departure time.Time,
	arrival time.Time,
	packetSizeBit uint64,
	marked bool,
) {
	sentTs := uint64(departure.UnixMicro())
	recvTs := uint64(arrival.UnixMicro())

	// current one-way delay (d_fwd)
	oneWayDelay := recvTs - sentTs

	// update base delay
	oldBase := r.baseDelay
	r.baseDelay = min(r.baseDelay, oneWayDelay)

	// TODO: recompute base delay from time to time e.g. 10min

	// update queue delay; only if new measurement
	if recvTs >= r.lastArraival {
		r.lastArraival = recvTs
		currDelay := oneWayDelay - r.baseDelay
		r.addNewDelay(currDelay)

	} else if r.baseDelay != oldBase {
		// we skipped qdelay calculation
		// but the base delay has changed
		updatedDelay := r.qDelay + oldBase - r.baseDelay
		r.addNewDelay(updatedDelay)
	}

	// check for queue build-up
	queueBuildup := false
	if r.qDelay >= r.config.QEPS {
		queueBuildup = true
	}

	// update logwin
	r.logWin.NewMediaPacketRecieved(packetNumber, recvTs, packetSizeBit, marked, queueBuildup)
	r.logWin.UpdateStats(recvTs)

	// update reciving rate
	// TODO: this might overflow
	recvBitsSeconds := r.logWin.ReceivedBits() * 1000000 // to convert micro seconds to seconds

	r.recvRate = recvBitsSeconds / r.config.LogWin
}

func (r *Receiver) addNewDelay(newDelay uint64) {
	// filter qdelay with min filter
	r.delayWin.AddSample(newDelay)
	minDelay := r.delayWin.MinDelay()

	r.qDelay = SmoothDelaySamples(*r.config, minDelay, r.qDelay)
}

// GenerateFeedback for loss detection at sender.
// On time to send a new feedback report (t_curr - t_last > DELTA).
func (r *Receiver) GenerateFeedback() (feedback Feedback) {
	feedback.RecvRate = r.recvRate
	feedback.Delay = r.qDelay
	feedback.QueueBuildup = r.logWin.QueueBuildup()

	return feedback
}
