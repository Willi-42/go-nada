package nada

const (
	PRIO      float64 = 1.0
	RMIN      uint64  = 150000   // bps
	RMAX      uint64  = 15000000 // bps
	XREF      uint64  = 10       // ms
	KAPPA     float64 = 0.5
	ETA       float64 = 2.0
	TAU       uint64  = 500 // ms
	DELTA     uint64  = 100 // ms
	LOGWIN    uint64  = 500 // ms
	QEPS      uint64  = 10  // ms
	DFILT     uint64  = 120 // ms
	GAMMA_MAX float64 = 0.5
	QBOUND    uint64  = 50 // ms
	MULTILOSS float64 = 7.0
	QTH       uint64  = 7.0
	LAMBDA    float64 = 0.5
	PLRREF    float64 = 0.01
	PMRREF    float64 = 0.01
	DLOSS     uint64  = 10 // ms
	DMARK     uint64  = 2  // ms
	FPS       uint64  = 30
	BETA_S    float64 = 0.1
	BETA_V    float64 = 0.1
	ALPHA     float64 = 0.1
)

type Config struct {
	Priority      float64 // Weight of priority of the flow
	MinRate       uint64  // minimum rate supported by encoder in bps
	MaxRate       uint64  // maximum rate supported by encoder in bps
	RefCongLevel  uint64  // Reference congestion level
	Kappa         float64 // Scaling parameter for gradual rate update calculation
	Eta           float64 // Scaling parameter for gradual rate update calculation
	Tau           uint64  // Upper bound of RTT in gradual rate update calculation
	FeedbackDelta uint64  // Target feedback interval
	LogWin        uint64  // Observation window in time for calculating packet summary statistics at receiver
	QEPS          uint64  // Threshold for determining queuing delay buildup at receiver
	DFILT         uint64  // Bound on filtering delay
	GAMMA_MAX     float64 // Upper bound on rate increase ratio for accelerated ramp up
	QBOUND        uint64  // Upper bound on self-inflicted queuing delay during ramp up
	MULTILOSS     float64 // Multiplier for self-scaling the expiration threshold of the last observed loss (loss_exp) based on measured average loss interval (loss_int)
	QTH           uint64  // Delay threshold for invoking non-linear warping
	LAMBDA        float64 // Scaling parameter in the exponent of non-linear warping
	PLRREF        float64 // Reference packet loss ratio
	PMRREF        float64 // Reference packet marking ratio
	DLOSS         uint64  // Reference delay penalty for loss when packet loss ratio is at PLRREF
	DMARK         uint64  // Reference delay penalty for ECN marking when packet marking is at PM
	FPS           uint64  // Frame rate of incoming video
	BETA_S        float64 // Scaling parameter for modulating outgoing sending rate
	BETA_V        float64 // Scaling parameter for modulating video encoder target rate
	ALPHA         float64 // Smoothing factor in exponential smoothing of packet loss and marking ratios
}

func populateConfig(c *Config) *Config {
	if c.Priority == 0.0 {
		c.Priority = PRIO
	}

	c.MinRate = setDefaultInt(c.MinRate, RMIN)
	c.MaxRate = setDefaultInt(c.MaxRate, RMAX)
	c.RefCongLevel = setDefaultInt(c.RefCongLevel, XREF)
	c.Kappa = setDefaultFloat(c.Kappa, KAPPA)
	c.Eta = setDefaultFloat(c.Eta, ETA)
	c.Tau = setDefaultInt(c.Tau, TAU)
	c.FeedbackDelta = setDefaultInt(c.FeedbackDelta, DELTA)
	c.LogWin = setDefaultInt(c.LogWin, LOGWIN)
	c.QEPS = setDefaultInt(c.QEPS, QEPS)
	c.DFILT = setDefaultInt(c.DFILT, DFILT)
	c.GAMMA_MAX = setDefaultFloat(c.GAMMA_MAX, GAMMA_MAX)
	c.QBOUND = setDefaultInt(c.QBOUND, QBOUND)
	c.MULTILOSS = setDefaultFloat(c.MULTILOSS, MULTILOSS)
	c.QTH = setDefaultInt(c.QTH, QTH)
	c.LAMBDA = setDefaultFloat(c.LAMBDA, LAMBDA)
	c.PLRREF = setDefaultFloat(c.PLRREF, PLRREF)
	c.PMRREF = setDefaultFloat(c.PMRREF, PMRREF)
	c.DLOSS = setDefaultInt(c.DLOSS, DLOSS)
	c.DMARK = setDefaultInt(c.DMARK, DMARK)
	c.FPS = setDefaultInt(c.FPS, FPS)

	c.BETA_S = setDefaultFloat(c.BETA_S, BETA_S)
	c.BETA_V = setDefaultFloat(c.BETA_V, BETA_V)
	c.ALPHA = setDefaultFloat(c.ALPHA, ALPHA)

	return c
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
