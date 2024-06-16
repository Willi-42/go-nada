package nada

import (
	"math"
	"time"
)

// Reciver: all timestamps are in microseconds
type Receiver struct {
	t_curr  uint64 // current time
	t_last  uint64 // last time sending a feedback message
	d_base  uint64 // estimated baseline delay
	d_tilde uint64 // Equivalent delay after non-linear warping
	d_queue uint64 // estimated queuing delay
	r_recv  uint64 // receiving rate

	// Threshold value for setting the last observed packet loss to expiration.
	// Measured in terms of packet counts.
	loss_exp uint64

	x_curr uint64 // Aggregate congestion signal

	// packet stats
	p_loss uint64 // estimated packet loss ratio
	p_mark uint64 // estimated packet ECN marking ratio

	config    *Config
	logWindow *logWinQueue
}

func NewReceiver(config Config) Receiver {
	currTime := time.Now().UnixMilli()

	configPopulated := populateConfig(&config)
	logWinSize := configPopulated.LogWin * 1000 // convert to micro sec

	return Receiver{
		d_base:    math.MaxUint64, // set to infinity
		p_loss:    0,
		p_mark:    0,
		r_recv:    0,
		t_last:    uint64(currTime),
		t_curr:    uint64(currTime),
		config:    configPopulated,
		logWindow: NewLogWinQueue(logWinSize),
	}
}

func (r *Receiver) PacketArrived(pn uint64, sentTs uint64, recvTs uint64, packetSize uint64, marked bool) {
	r.t_curr = recvTs // TODO: should this be current time instead?

	d_fwd := recvTs - sentTs // one-way delay

	// update base delay
	if r.d_base > d_fwd {
		r.d_base = d_fwd
	}
	// TODO: recompute base delay from time to time e.g. 10min

	// update queue delay
	r.d_queue = d_fwd - r.d_base
	// TODO: add min filter to d_queue
	// compare: https://www.rfc-editor.org/rfc/rfc8698#section-5.1.1

	// check for queue build-up
	queueBuildup := false
	if r.d_base >= r.config.QEPS {
		queueBuildup = true
	}

	// update logwin
	r.logWindow.NewMediaPacketRecieved(pn, recvTs, packetSize, marked, queueBuildup)
	r.logWindow.updateStats(recvTs)
	// TODO: calc loss_int

	// calculate loss/marking ratio
	totoalPackets := r.logWindow.numberPacketArrived + r.logWindow.numberLostPackets
	r.p_loss = r.calcSmoothedRatio(r.logWindow.numberLostPackets, totoalPackets, r.p_loss)
	r.p_mark = r.calcSmoothedRatio(r.logWindow.numberMarkedPackets, totoalPackets, r.p_mark)

	// update reciving rate
	// recived_bytes_in_logwin / logWintimeInMs * 1000 = bps
	r.r_recv = uint64((float64(r.logWindow.accumulatedSize) / float64(r.config.LogWin)) * 1000)
}

// GenerateFeedback: On time to send a new feedback report (t_curr - t_last > DELTA)
func (r *Receiver) GenerateFeedback() (uint64, uint64, bool) {

	// loss_exp is configured to self-scale with the average packet loss
	// interval loss_int with a multiplier MULTILOSS
	loss_int := r.logWindow.lossInt.calcAvgLossInt()
	r.loss_exp = uint64(r.config.MULTILOSS * loss_int)

	// calculate non-linear warping of delay d_tilde if packet loss exists
	// Only if the last observed packet loss is within the expiration
	// window of loss_exp (measured in terms of packet counts), we apply non-lin warapping
	if r.logWindow.numberOfPacketsSinceLastLoss <= r.loss_exp {
		r.d_tilde = r.calcNonLinWrappingQDelay()
	} else {
		r.d_tilde = r.d_queue
	}

	// calculate current aggregate congestion signal x_curr
	r.x_curr = r.calcAggregateCng()

	// determine mode of rate adaptation for sender: rmode
	// if packetloss in logwin and no queue build up
	// for all previous delay samples within the observation window LOGWIN
	rampUpMode := false
	if r.logWindow.numberLostPackets == 0 && r.logWindow.numberQueueBuildup == 0 {
		rampUpMode = true
	}

	// send feedback containing values of: rmode, x_curr, and r_recv
	// TODO: update t_last = t_curr

	return r.r_recv, r.x_curr, rampUpMode
}
