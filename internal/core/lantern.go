package core

import (
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

type Lantern struct {
	Position  utils.Vector2D
	Radius    float64
	Intensity float64
	PulsePhase float64
}

func NewLantern(x, y float64) *Lantern {
	return &Lantern{
		Position:  utils.Vector2D{X: x, Y: y},
		Radius:    config.LanternRadius,
		Intensity: 1.0,
		PulsePhase: 0.0,
	}
}

func (l *Lantern) Update(dt float64) {
	l.PulsePhase += dt * 2.0
	if l.PulsePhase > 1.0 {
		l.PulsePhase = 0.0
	}
}

func (l *Lantern) GetIntensity() float64 {
	return 0.7 + 0.3*l.PulsePhase
}