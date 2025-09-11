package nada

const (
	PRIO               float64 = 1.0
	RMIN               uint64  = 150000  // bps
	RMAX               uint64  = 1500000 // bps
	XREF               uint64  = 10      // ms
	KAPPA              float64 = 0.5
	ETA                float64 = 2.0
	TAU                uint64  = 500 // ms
	DELTA              uint64  = 100 // ms
	LOGWIN             uint64  = 500 // ms
	QEPS               uint64  = 10  // ms
	DFILT              uint64  = 120 // ms
	GAMMA_MAX          float64 = 0.1
	QBOUND             uint64  = 50 // ms
	MULTILOSS          float64 = 7.0
	QTH                uint64  = 50 // ms
	LAMBDA             float64 = 0.5
	PLRREF             float64 = 0.01
	PMRREF             float64 = 0.01
	DLOSS              uint64  = 10 // ms
	DMARK              uint64  = 2  // ms
	ALPHA              float64 = 0.1
	GRAD_UPDATE_FACTOR float64 = 0.02
)

type Config struct {
	Priority               float64 // Weight of priority of the flow
	MinRate                uint64  // minimum rate supported by encoder in bps
	MaxRate                uint64  // maximum rate supported by encoder in bps
	StartRate              uint64  // at which rate NADA should start (default: MinRate)
	RefCongLevel           uint64  // Reference congestion level
	Kappa                  float64 // Scaling parameter for gradual rate update calculation
	Eta                    float64 // Scaling parameter for gradual rate update calculation
	Tau                    uint64  // Upper bound of RTT in gradual rate update calculation
	FeedbackDelta          uint64  // Target feedback interval
	LogWin                 uint64  // Observation window in time for calculating packet summary statistics at receiver
	QEPS                   uint64  // Threshold for determining queuing delay buildup at receiver
	DFILT                  uint64  // Bound on filtering delay for RampUp Mode
	MaxRampUpFactor        float64 // GAMMA_MAX: Upper bound on rate increase ratio for accelerated ramp up
	MaxGradualUpdateFactor float64 // Upper/Lower bound on rate increase/decrease ratio for gradual updates
	QBOUND                 uint64  // Upper bound on self-inflicted queuing delay during ramp up
	MULTILOSS              float64 // Multiplier for self-scaling the expiration threshold of the last observed loss (loss_exp) based on measured average loss interval (loss_int)
	QTH                    uint64  // Delay threshold for invoking non-linear warping
	LAMBDA                 float64 // Scaling parameter in the exponent of non-linear warping
	PLRREF                 float64 // Reference packet loss ratio
	PMRREF                 float64 // Reference packet marking ratio
	DLOSS                  uint64  // Reference delay penalty for loss when packet loss ratio is at PLRREF
	DMARK                  uint64  // Reference delay penalty for ECN marking when packet marking is at PM
	ALPHA                  float64 // Smoothing factor in exponential smoothing of packet loss and marking ratios

	DeactivateQDelayWrapping bool // do not apply wrapping of qdely
	SmoothDelaySamples       bool // Smooth delay samples with exponential moving average
	UseDefaultGradualUpdates bool // Use default gradual update mode that depends on MaxRate
}

func populateConfig(c *Config) *Config {
	if c.Priority == 0.0 {
		c.Priority = PRIO
	}

	c.MinRate = setDefaultInt(c.MinRate, RMIN)
	c.MaxRate = setDefaultInt(c.MaxRate, RMAX)
	c.StartRate = setDefaultInt(c.StartRate, RMIN)
	c.RefCongLevel = setDefaultMs(c.RefCongLevel, XREF)
	c.Kappa = setDefaultFloat(c.Kappa, KAPPA)
	c.Eta = setDefaultFloat(c.Eta, ETA)
	c.Tau = setDefaultMs(c.Tau, TAU)
	c.FeedbackDelta = setDefaultMs(c.FeedbackDelta, DELTA)
	c.LogWin = setDefaultMs(c.LogWin, LOGWIN)
	c.QEPS = setDefaultMs(c.QEPS, QEPS)
	c.DFILT = setDefaultMs(c.DFILT, DFILT)
	c.MaxRampUpFactor = setDefaultFloat(c.MaxRampUpFactor, GAMMA_MAX)
	c.MaxGradualUpdateFactor = setDefaultFloat(c.MaxGradualUpdateFactor, GRAD_UPDATE_FACTOR)
	c.QBOUND = setDefaultMs(c.QBOUND, QBOUND)
	c.MULTILOSS = setDefaultFloat(c.MULTILOSS, MULTILOSS)
	c.QTH = setDefaultMs(c.QTH, QTH)
	c.LAMBDA = setDefaultFloat(c.LAMBDA, LAMBDA)
	c.PLRREF = setDefaultFloat(c.PLRREF, PLRREF)
	c.PMRREF = setDefaultFloat(c.PMRREF, PMRREF)
	c.DLOSS = setDefaultMs(c.DLOSS, DLOSS)
	c.DMARK = setDefaultMs(c.DMARK, DMARK)
	c.ALPHA = setDefaultFloat(c.ALPHA, ALPHA)

	// UseDefaultGradualUpdates, DeactivateQDelayWrapping & SmoothDelaySamples are false if not given == default

	return c
}

// setDefaultMs sets default and converts to micro seconds
func setDefaultMs(value, defaultValue uint64) uint64 {
	return setDefaultInt(value, defaultValue) * 1000
}

func setDefaultInt(value, defaultValue uint64) uint64 {
	if value == 0 {
		return defaultValue
	}

	return value
}

func setDefaultFloat(value, defaultValue float64) float64 {
	if value == 0.0 {
		return defaultValue
	}

	return value
}
