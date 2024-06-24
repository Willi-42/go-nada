package nada

import (
	"math"
)

// smoothedRatio calculates the smoothed loss/marking ratio
func smoothedRatio(conf Config, currentCnt, totoalCnt, prevRatio uint64) uint64 {
	currRatio := float64(currentCnt) / float64(totoalCnt)
	return uint64(conf.ALPHA*currRatio + (1-conf.ALPHA)*float64(prevRatio))
}

// aggregateCng calculates the aggregated congestion signal (x_curr)
func aggregateCng(conf Config, d_tilde, p_mark, p_loss uint64) uint64 {
	dmark := float64(conf.DMARK) * math.Pow(float64(p_mark)/conf.PMRREF, 2)
	dloss := float64(conf.DLOSS) * math.Pow(float64(p_loss)/conf.PLRREF, 2)
	return d_tilde + uint64(dmark+dloss)
}

// nonLinWrappingQDelay calculates the non linear wrapping of the queueing dleay
func nonLinWrappingQDelay(conf Config, d_queue uint64) uint64 {
	if d_queue < conf.QTH {
		return d_queue
	} else {
		tmp := -conf.LAMBDA * (float64(d_queue-conf.QTH) / float64(conf.QTH))
		return conf.QTH * uint64(math.Exp(tmp))
	}
}

// rampUpRate calculates the reference rate in rampUp mode
func rampUpRate(
	config Config,
	rtt uint64,
	prevRefRate uint64,
	recvRate uint64,
) uint64 {

	bound := float64(config.QBOUND) / float64(rtt+config.FeedbackDelta+config.DFILT)
	gamma := min(config.GAMMA_MAX, bound)

	incrRecvRate := (1 + gamma) * float64(recvRate)

	return max(prevRefRate, uint64(incrRecvRate))
}

// gradualUpdateRate calculates the reference rate in gradual update mode
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
