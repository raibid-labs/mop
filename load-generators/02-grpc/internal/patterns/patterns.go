package patterns

import (
	"math"
	"time"
)

// LoadPattern defines the interface for different load patterns
type LoadPattern interface {
	RPS(second int) int
}

// ConstantLoad generates constant requests per second
type ConstantLoad struct {
	rps int
}

func NewConstantLoad(rps int) *ConstantLoad {
	return &ConstantLoad{rps: rps}
}

func (c *ConstantLoad) RPS(second int) int {
	return c.rps
}

// SpikeLoad generates normal load with periodic spikes
type SpikeLoad struct {
	baseRPS    int
	spikeRPS   int
	spikeEvery time.Duration
}

func NewSpikeLoad(baseRPS, spikeRPS int, spikeEvery time.Duration) *SpikeLoad {
	return &SpikeLoad{
		baseRPS:    baseRPS,
		spikeRPS:   spikeRPS,
		spikeEvery: spikeEvery,
	}
}

func (s *SpikeLoad) RPS(second int) int {
	spikeDuration := int(s.spikeEvery.Seconds())
	if spikeDuration == 0 {
		spikeDuration = 10
	}

	// Spike every N seconds for 3 seconds
	if second%spikeDuration < 3 {
		return s.spikeRPS
	}
	return s.baseRPS
}

// RampLoad gradually increases load from min to max over duration
type RampLoad struct {
	minRPS   int
	maxRPS   int
	duration time.Duration
}

func NewRampLoad(minRPS, maxRPS int, duration time.Duration) *RampLoad {
	return &RampLoad{
		minRPS:   minRPS,
		maxRPS:   maxRPS,
		duration: duration,
	}
}

func (r *RampLoad) RPS(second int) int {
	totalSeconds := int(r.duration.Seconds())
	if totalSeconds == 0 {
		totalSeconds = 60
	}

	if second >= totalSeconds {
		return r.maxRPS
	}

	progress := float64(second) / float64(totalSeconds)
	rps := float64(r.minRPS) + (float64(r.maxRPS-r.minRPS) * progress)
	return int(math.Round(rps))
}

// StepLoad increases load in steps
type StepLoad struct {
	startRPS  int
	stepRPS   int
	stepEvery time.Duration
}

func NewStepLoad(startRPS, stepRPS int, stepEvery time.Duration) *StepLoad {
	return &StepLoad{
		startRPS:  startRPS,
		stepRPS:   stepRPS,
		stepEvery: stepEvery,
	}
}

func (s *StepLoad) RPS(second int) int {
	stepDuration := int(s.stepEvery.Seconds())
	if stepDuration == 0 {
		stepDuration = 10
	}

	step := second / stepDuration
	return s.startRPS + (step * s.stepRPS)
}

// WaveLoad generates sinusoidal wave pattern
type WaveLoad struct {
	minRPS int
	maxRPS int
	period time.Duration
}

func NewWaveLoad(minRPS, maxRPS int, period time.Duration) *WaveLoad {
	return &WaveLoad{
		minRPS: minRPS,
		maxRPS: maxRPS,
		period: period,
	}
}

func (w *WaveLoad) RPS(second int) int {
	periodSeconds := w.period.Seconds()
	if periodSeconds == 0 {
		periodSeconds = 60
	}

	// Calculate position in wave (0 to 2π)
	position := 2 * math.Pi * float64(second) / periodSeconds

	// Sin wave oscillates between -1 and 1
	amplitude := float64(w.maxRPS-w.minRPS) / 2
	midpoint := float64(w.minRPS+w.maxRPS) / 2

	rps := midpoint + amplitude*math.Sin(position)
	return int(math.Round(rps))
}
