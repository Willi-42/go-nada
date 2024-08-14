package nada

import (
	"math"

	"github.com/Willi-42/go-nada/nada/windows"
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

// nonLinWrapingQDelay calculates the non linear wrapping (d_tilde) of the queueing delay (d_queue)
func nonLinWrapingQDelay(conf Config, qDelay uint64) uint64 {
	exponent := -conf.LAMBDA * (float64(qDelay-conf.QTH) / float64(conf.QTH))
	return uint64(float64(conf.QTH) * math.Exp(exponent))
}

func wrapQDelay(conf Config, qDelay uint64, logWin *windows.LogWindow) uint64 {
	updatedDelay := qDelay

	if conf.DeactivateQDelayWrapping {
		return updatedDelay
	}

	// loss_exp self-scales with the average packet loss interval with a multiplier MULTILOSS
	// Threshold value for setting the last observed packet loss to expiration.
	// Measured in terms of packet counts.
	avgLossInt := logWin.AvgLossInterval()
	lossExp := uint64(conf.MULTILOSS * avgLossInt)

	// calculate non-linear warping of delay (d_tilde)
	// if the last observed packet loss is within the expiration window of loss_exp

	packtesSinceLoss, gotLoss := logWin.PacketsSinceLoss()

	if gotLoss && packtesSinceLoss <= lossExp && qDelay >= conf.QTH {
		updatedDelay = nonLinWrapingQDelay(conf, qDelay)
	}

	return updatedDelay
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
	if prevRefRate == 0 {
		prevRefRate = 1 // prevent division by 0
	}

	// offset to the ideal congestion
	xIdeal := conf.Priority * float64(conf.RefCongLevel) * (float64(conf.MaxRate) / float64(prevRefRate))
	xOffset := float64(xCurr) - xIdeal

	// current congestion signal change
	xDiff := int64(xCurr) - int64(xPrev)

	term1 := conf.Kappa * (float64(feedbackDelta) / float64(conf.Tau))
	term1 *= (xOffset / float64(conf.Tau)) * float64(prevRefRate)

	term2 := conf.Kappa * conf.Eta * (float64(xDiff) / float64(conf.Tau)) * float64(prevRefRate)

	res := int64(prevRefRate) - int64(term1) - int64(term2)

	// TODO: why can there be a negative result
	if res < 0 {
		res = 0
	}

	return uint64(res)
}
