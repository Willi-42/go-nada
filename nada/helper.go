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
