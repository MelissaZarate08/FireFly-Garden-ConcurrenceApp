package core

import (
	"context"
	"math"
	"time"

	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

type WindDirection int

const (
	WindNone WindDirection = iota
	WindNorth
	WindSouth
	WindEast
	WindWest
	WindNorthEast
	WindNorthWest
	WindSouthEast
	WindSouthWest
)

type Wind struct {
	direction WindDirection
	force     utils.Vector2D
	strength  float64
}

func NewWind() *Wind {
	return &Wind{
		direction: WindEast,
		strength:  config.WindForce,
		force:     utils.Vector2D{X: config.WindForce, Y: 0},
	}
}

func (w *Wind) Run(ctx context.Context) {
	ticker := time.NewTicker(config.WindChangeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
			
		case <-ticker.C:
			w.changeDirection()
		}
	}
}

func (w *Wind) changeDirection() {
	directions := []WindDirection{
		WindNorth, WindSouth, WindEast, WindWest,
		WindNorthEast, WindNorthWest, WindSouthEast, WindSouthWest,
	}
	
	w.direction = directions[int(utils.RandomFloat(0, float64(len(directions))))]
	w.updateForce()
}

func (w *Wind) SetDirection(dir WindDirection) {
	w.direction = dir
	w.updateForce()
}

func (w *Wind) GetDirection() WindDirection {
	return w.direction
}

func (w *Wind) GetForce() utils.Vector2D {
	return w.force
}

func (w *Wind) GetForcePointer() *utils.Vector2D {
	return &w.force
}

func (w *Wind) updateForce() {
	angle := w.directionToAngle()
	w.force = utils.Vector2D{
		X: math.Cos(angle) * w.strength,
		Y: math.Sin(angle) * w.strength,
	}
}

func (w *Wind) directionToAngle() float64 {
	switch w.direction {
	case WindNorth:
		return -math.Pi / 2
	case WindSouth:
		return math.Pi / 2
	case WindEast:
		return 0
	case WindWest:
		return math.Pi
	case WindNorthEast:
		return -math.Pi / 4
	case WindNorthWest:
		return -3 * math.Pi / 4
	case WindSouthEast:
		return math.Pi / 4
	case WindSouthWest:
		return 3 * math.Pi / 4
	default:
		return 0
	}
}

func (w *Wind) GetDirectionName() string {
	switch w.direction {
	case WindNone:
		return "None"
	case WindNorth:
		return "North"
	case WindSouth:
		return "South"
	case WindEast:
		return "East"
	case WindWest:
		return "West"
	case WindNorthEast:
		return "NorthEast"
	case WindNorthWest:
		return "NorthWest"
	case WindSouthEast:
		return "SouthEast"
	case WindSouthWest:
		return "SouthWest"
	default:
		return "Unknown"
	}
}

func (w *Wind) CycleDirection() {
	directions := []WindDirection{
		WindNorth, WindNorthEast, WindEast, WindSouthEast,
		WindSouth, WindSouthWest, WindWest, WindNorthWest,
	}
	
	currentIndex := -1
	for i, dir := range directions {
		if dir == w.direction {
			currentIndex = i
			break
		}
	}
	
	nextIndex := (currentIndex + 1) % len(directions)
	w.SetDirection(directions[nextIndex])
}