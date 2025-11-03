package core

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

type FireflyState struct {
	ID         int
	Position   utils.Vector2D
	Brightness float64
	IsAlive    bool
}

var droppedStates uint64

func GetDroppedStates() uint64 {
	return atomic.LoadUint64(&droppedStates)
}

type Firefly struct {
	id              int
	position        utils.Vector2D
	velocity        utils.Vector2D
	brightness      float64
	blinkPhase      float64
	blinkCycleDur   float64
	targetPosition  *utils.Vector2D
	attractionPoint *utils.Vector2D
	windForce       *utils.Vector2D

	age      float64
	lifespan float64
}

func NewFirefly(id int, spawnX, spawnY float64) *Firefly {
	return &Firefly{
		id:            id,
		position:      utils.Vector2D{X: spawnX, Y: spawnY},
		velocity:      utils.RandomUnitVector().Mul(config.FireflySpeed),
		brightness:    0.0,
		blinkPhase:    0.0,
		blinkCycleDur: utils.RandomFloat(config.FireflyBlinkCycleMin, config.FireflyBlinkCycleMax),
		age:           0.0,
		lifespan:      utils.RandomFloat(config.FireflyLifespanMin, config.FireflyLifespanMax),
	}
}

func (f *Firefly) Run(ctx context.Context, stateCh chan<- FireflyState, lanterns []*Lantern, dt float64) {
	ticker := time.NewTicker(time.Second / time.Duration(config.TargetFPS))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			f.publishState(stateCh, false)
			return

		case <-ticker.C:
			f.update(lanterns, dt)

			f.age += dt
			if f.age > f.lifespan {
				f.publishState(stateCh, false)
				return
			}

			f.publishState(stateCh, true)
		}
	}
}


func (f *Firefly) publishState(stateCh chan<- FireflyState, isAlive bool) {
	state := FireflyState{
		ID:         f.id,
		Position:   f.position,
		Brightness: f.brightness,
		IsAlive:    isAlive,
	}

	select {
	case stateCh <- state:
	default:
		atomic.AddUint64(&droppedStates, 1)
	}
}

func (f *Firefly) update(lanterns []*Lantern, dt float64) {
	f.updateBlinkPhase(dt)

	f.applyWandering()
	f.applyLanternAttraction(lanterns)
	f.applyAttractionPoint()
	f.applyWind()

	f.position = f.position.Add(f.velocity.Mul(dt))

	f.position = utils.WrapAround(f.position, config.ScreenWidth, config.ScreenHeight)

	if f.velocity.Magnitude() > config.FireflySpeed*2 {
		f.velocity = f.velocity.Normalize().Mul(config.FireflySpeed * 2)
	}
}

func (f *Firefly) updateBlinkPhase(dt float64) {
	f.blinkPhase += dt / f.blinkCycleDur
	if f.blinkPhase > 1.0 {
		f.blinkPhase = 0.0
		f.blinkCycleDur = utils.RandomFloat(config.FireflyBlinkCycleMin, config.FireflyBlinkCycleMax)
	}

	f.brightness = (math.Sin(f.blinkPhase*2*math.Pi) + 1) / 2
}

func (f *Firefly) applyWandering() {
	if utils.RandomFloat(0, 1) < 0.05 { 
		randomForce := utils.RandomUnitVector().Mul(0.2)
		f.velocity = f.velocity.Add(randomForce)
	}
}

func (f *Firefly) applyLanternAttraction(lanterns []*Lantern) {
	if lanterns == nil || len(lanterns) == 0 {
		return
	}

	for _, lantern := range lanterns {
		distance := utils.Distance(f.position, lantern.Position)

		if distance < lantern.Radius && distance > 1 {
			direction := lantern.Position.Sub(f.position).Normalize()

			strength := (lantern.Radius - distance) / lantern.Radius
			force := direction.Mul(config.LanternInfluenceForce * strength)

			f.velocity = f.velocity.Add(force)
		}
	}
}

func (f *Firefly) applyAttractionPoint() {
	if f.attractionPoint == nil {
		return
	}

	distance := utils.Distance(f.position, *f.attractionPoint)

	if distance > 10 {
		direction := f.attractionPoint.Sub(f.position).Normalize()
		force := direction.Mul(config.FireflyAttractionForce)
		f.velocity = f.velocity.Add(force)
	}
}

func (f *Firefly) applyWind() {
	if f.windForce == nil {
		return
	}

	windEffect := f.windForce.Mul(config.FireflyWindResistance)
	f.velocity = f.velocity.Add(windEffect)
}

func (f *Firefly) SetAttractionPoint(point *utils.Vector2D) {
	f.attractionPoint = point
}

func (f *Firefly) SetWindForce(wind *utils.Vector2D) {
	f.windForce = wind
}

func (f *Firefly) GetID() int {
	return f.id
}
