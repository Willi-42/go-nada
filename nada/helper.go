package nada

import "math"

// calcSmoothedRatio calculates the smoothed loss/marking ratio
func (r *Receiver) calcSmoothedRatio(currentCnt, totoalCnt, prevRatio uint64) uint64 {
	currRatio := float64(currentCnt) / float64(totoalCnt)
	return uint64(r.config.ALPHA*currRatio + (1-r.config.ALPHA)*float64(prevRatio))
}

func (r *Receiver) calcAggregateCng() uint64 {
	dmark := float64(r.config.DMARK) * math.Pow(float64(r.p_mark)/r.config.PMRREF, 2)
	dloss := float64(r.config.DLOSS) * math.Pow(float64(r.p_loss)/r.config.PLRREF, 2)
	return r.d_tilde + uint64(dmark+dloss)
}

func (r *Receiver) calcNonLinWrappingQDelay() uint64 {
	if r.d_queue < r.config.QTH {
		return r.d_queue
	} else {
		tmp := -r.config.LAMBDA * (float64(r.d_queue-r.config.QTH) / float64(r.config.QTH))
		return uint64(math.Exp(tmp))
	}
}

func rampUpRate(config Config, rtt uint64, prevRefRate uint64, recvRate uint64) uint64 {
	bound := float64(config.QBOUND) / float64(rtt+config.FeedbackDelta+config.DFILT)
	gamma := min(config.GAMMA_MAX, bound)

	incrRecvRate := (1 + gamma) * float64(recvRate)

	return max(prevRefRate, uint64(incrRecvRate))
}

func gradualUpdateRate(
	conf Config,
	prevRefRate uint64,
	xCurr uint64,
	xPrev uint64,
	feedbackDelta uint64,
) uint64 {
	tmp := conf.Priority * float64(conf.RefCongLevel) * float64(conf.MaxRate) / float64(prevRefRate)
	xOffset := float64(xCurr) - tmp

	xDiff := xCurr - xPrev

	calc1 := conf.Kappa * (float64(feedbackDelta) / float64(conf.Tau))
	calc1 *= xOffset / float64(conf.Tau) * float64(prevRefRate)

	calc2 := conf.Kappa * conf.Eta * (float64(xDiff) / float64(conf.Tau)) * float64(prevRefRate)

	return prevRefRate - uint64(calc1) - uint64(calc2)

}
