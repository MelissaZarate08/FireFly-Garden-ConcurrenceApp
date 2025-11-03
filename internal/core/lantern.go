package core

import (
	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

// Lantern representa un farol que atrae luciérnagas
// Los faroles son estáticos y se crean mediante input del usuario
type Lantern struct {
	Position  utils.Vector2D
	Radius    float64
	Intensity float64
	PulsePhase float64
}

// NewLantern crea un nuevo farol en la posición especificada
func NewLantern(x, y float64) *Lantern {
	return &Lantern{
		Position:  utils.Vector2D{X: x, Y: y},
		Radius:    config.LanternRadius,
		Intensity: 1.0,
		PulsePhase: 0.0,
	}
}

// Update actualiza el estado del farol (animación de pulso)
func (l *Lantern) Update(dt float64) {
	// Ciclo de pulso para efecto visual
	l.PulsePhase += dt * 2.0
	if l.PulsePhase > 1.0 {
		l.PulsePhase = 0.0
	}
}

// GetIntensity devuelve la intensidad actual considerando el pulso
func (l *Lantern) GetIntensity() float64 {
	// Intensidad varía entre 0.7 y 1.0
	return 0.7 + 0.3*l.PulsePhase
}