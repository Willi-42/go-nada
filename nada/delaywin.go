package nada

import "slices"

type delayWindow struct {
	delaySampls []uint64
	size        int
}

func newDelayWin(size int) *delayWindow {
	return &delayWindow{
		delaySampls: make([]uint64, 0),
		size:        size,
	}
}

// addSample adds a delay sample to the window
func (d *delayWindow) addSample(delay uint64) {

	d.delaySampls = append(d.delaySampls, delay)

	// drop oldest interval
	if len(d.delaySampls) >= d.size {
		d.delaySampls = d.delaySampls[1:]
	}
}

// minDelay returns the minimum delay in the current window
func (d *delayWindow) minDelay() uint64 {
	return slices.Min(d.delaySampls)
}
