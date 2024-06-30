package windows

// Loss interval calculation based on https://www.rfc-editor.org/rfc/rfc5348#section-5.4
// However this start a new loss interval with every loss.
// TODO: should this be changed to TRFC (rfc5348) style loss interval (based on RTTs)
type lossInterval struct {
	intervals []uint64
	maxsize   int
	weigts    []float64 // for avg calc
}

func newLossIntervall(maxsize int) *lossInterval {
	weigts := make([]float64, maxsize)

	for i := range weigts {
		if i < maxsize/2 {
			weigts[i] = 1
			continue
		}
		weigts[i] = 2 * float64(maxsize-i) / float64(maxsize+2)
	}

	// maxSize is the size of the interval
	// We keep an additonal interval so the avg calc can decide to use newest
	// interfall one or not
	maxsize += 1

	return &lossInterval{
		intervals: make([]uint64, 0),
		maxsize:   maxsize,
		weigts:    weigts,
	}
}

// addLoss creates new loss interval
func (l *lossInterval) addLoss(lossGap uint64) {

	l.intervals = append(l.intervals, lossGap)

	// drop oldest interval
	if len(l.intervals) >= l.maxsize {
		l.intervals = l.intervals[1:]
	}
}

// addPacket adds packet to current loss interval
func (l *lossInterval) addPacket() {
	// no loss yet, first loss will start the first interval
	if len(l.intervals) == 0 {
		return
	}

	l.intervals[0]++
}

// Measured average loss interval in packet count
func (l *lossInterval) calcAvgLossInt() float64 {
	i_tot0 := float64(0)
	i_tot1 := float64(0)
	w_tot := float64(0)

	size := len(l.intervals)

	if size == 0 {
		return 0
	}

	// edge case: only one loss event
	if size == 1 {
		return 1 / float64(l.intervals[0])
	}

	// default case
	for i := 0; i < size-1; i++ {
		i_tot0 += float64(l.intervals[i]) * l.weigts[i]
		w_tot += l.weigts[i]
	}

	for i := 1; i < size; i++ {
		i_tot1 += float64(l.intervals[i]) * l.weigts[i-1]
	}
	i_tot := max(i_tot0, i_tot1)
	i_mean := i_tot / w_tot

	return 1 / i_mean
}
