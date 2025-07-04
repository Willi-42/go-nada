package windows

import "slices"

type DelayWindow struct {
	delaySampls []uint64
	size        int
}

func NewDelayWindow(size int) *DelayWindow {
	return &DelayWindow{
		delaySampls: make([]uint64, 0),
		size:        size,
	}
}

// AddSample adds a delay sample to the window
func (d *DelayWindow) AddSample(delay uint64) {

	d.delaySampls = append(d.delaySampls, delay)

	// drop oldest sample
	if len(d.delaySampls) > d.size {
		d.delaySampls = d.delaySampls[1:]
	}
}

// MinDelay returns the minimum delay in the current window
// compare: https://www.rfc-editor.org/rfc/rfc8698.html#name-method-for-delay-loss-and-m
func (d *DelayWindow) MinDelay() uint64 {
	return slices.Min(d.delaySampls)
}
