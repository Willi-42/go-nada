package nada

import (
	"math"
)

// Reciver: all timestamps are in microseconds
type Receiver struct {
	d_base  uint64 // estimated baseline delay
	d_tilde uint64 // Equivalent delay after non-linear warping
	d_queue uint64 // estimated queuing delay
	r_recv  uint64 // receiving rate

	// packet stats
	p_loss uint64 // estimated packet loss ratio
	p_mark uint64 // estimated packet ECN marking ratio

	config    *Config
	logWindow *logWinQueue
}

func NewReceiver(config Config) Receiver {
	configPopulated := populateConfig(&config)
	logWinSize := configPopulated.LogWin * 1000 // convert to micro sec

	return Receiver{
		d_base:    math.MaxUint64, // set to infinity
		p_loss:    0,
		p_mark:    0,
		r_recv:    0,
		config:    configPopulated,
		logWindow: NewLogWinQueue(logWinSize),
	}
}

// PacketArrived registers a new arrived packet.
// sentTs and recvTs are timestamps in microseconds.
// packetSize is the size of the packet in bytes.
// marked: packet got ECN
func (r *Receiver) PacketArrived(
	packetNumber uint64,
	sentTs uint64,
	recvTs uint64,
	packetSize uint64,
	marked bool,
) {
	d_fwd := recvTs - sentTs // current one-way delay

	// update base delay
	r.d_base = min(r.d_base, d_fwd)

	// TODO: recompute base delay from time to time e.g. 10min

	// update queue delay
	r.d_queue = d_fwd - r.d_base
	// TODO: add min filter to d_queue
	// compare: https://www.rfc-editor.org/rfc/rfc8698.html#name-method-for-delay-loss-and-m

	// check for queue build-up
	queueBuildup := false
	if r.d_base >= r.config.QEPS {
		queueBuildup = true
	}

	// update logwin
	r.logWindow.NewMediaPacketRecieved(packetNumber, recvTs, packetSize, marked, queueBuildup)
	r.logWindow.updateStats(recvTs)

	// calculate loss/marking ratio
	totoalPackets := r.logWindow.numberPacketArrived + r.logWindow.numberLostPackets
	r.p_loss = r.calcSmoothedRatio(r.logWindow.numberLostPackets, totoalPackets, r.p_loss)
	r.p_mark = r.calcSmoothedRatio(r.logWindow.numberMarkedPackets, totoalPackets, r.p_mark)

	// update reciving rate
	// recived_bytes_in_logwin / logWintimeInMs * 1000 = bps
	r.r_recv = uint64((float64(r.logWindow.totalSize) / float64(r.config.LogWin)) * 1000)
}

// GenerateFeedback: On time to send a new feedback report (t_curr - t_last > DELTA)
func (r *Receiver) GenerateFeedback() (uint64, uint64, bool) {

	// loss_exp is configured to self-scale with the average packet loss
	// interval loss_int with a multiplier MULTILOSS
	// Threshold value for setting the last observed packet loss to expiration.
	// Measured in terms of packet counts.
	loss_int := r.logWindow.lossInt.calcAvgLossInt()
	loss_exp := uint64(r.config.MULTILOSS * loss_int)

	// calculate non-linear warping of delay d_tilde if packet loss exists
	// Only if the last observed packet loss is within the expiration
	// window of loss_exp (measured in terms of packet counts), we apply non-lin warapping
	if r.logWindow.numberPacketsSinceLoss <= loss_exp {
		r.d_tilde = r.calcNonLinWrappingQDelay()
	} else {
		r.d_tilde = r.d_queue
	}

	// calculate aggregate congestion signal x_curr
	x_curr := r.calcAggregateCng()

	// determine mode of rate adaptation for sender: rmode
	// if packetloss in logwin and no queue build up
	// for all previous delay samples within the observation window LOGWIN
	rampUpMode := false
	if r.logWindow.numberLostPackets == 0 && r.logWindow.queueBuildupCnt == 0 {
		rampUpMode = true
	}

	// send feedback containing values of: rmode, x_curr, and r_recv

	return r.r_recv, x_curr, rampUpMode
}
