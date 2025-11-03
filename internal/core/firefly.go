package core

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

// FireflyState representa el estado inmutable de una luciérnaga
type FireflyState struct {
	ID         int
	Position   utils.Vector2D
	Brightness float64 // 0.0 a 1.0
	IsAlive    bool
}

// contador de actualizaciones descartadas (atomic)
var droppedStates uint64

// GetDroppedStates devuelve el número de estados descartados
func GetDroppedStates() uint64 {
	return atomic.LoadUint64(&droppedStates)
}

// Firefly representa una luciérnaga con comportamiento autónomo
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

// NewFirefly crea una nueva luciérnaga con posición y velocidad aleatorias
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

// Run inicia la goroutine de la luciérnaga
// Actualiza su estado y lo publica al canal stateCh
// Termina cuando el contexto se cancela o la luciérnaga muere por edad
func (f *Firefly) Run(ctx context.Context, stateCh chan<- FireflyState, lanterns []*Lantern, dt float64) {
	ticker := time.NewTicker(time.Second / time.Duration(config.TargetFPS))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Publicar estado final antes de terminar
			f.publishState(stateCh, false)
			return

		case <-ticker.C:
			f.update(lanterns, dt)

			// Envejecer y morir si supera lifespan
			f.age += dt
			if f.age > f.lifespan {
				// publicar como muerta y salir
				f.publishState(stateCh, false)
				return
			}

			f.publishState(stateCh, true)
		}
	}
}

// publishState envía el estado actual al canal sin bloquear
// Si el canal está lleno, contabiliza el descarte en droppedStates
func (f *Firefly) publishState(stateCh chan<- FireflyState, isAlive bool) {
	state := FireflyState{
		ID:         f.id,
		Position:   f.position,
		Brightness: f.brightness,
		IsAlive:    isAlive,
	}

	// Non-blocking send
	select {
	case stateCh <- state:
	default:
		// Si el canal está lleno, contabilizamos el descarte (no bloqueamos)
		atomic.AddUint64(&droppedStates, 1)
	}
}

// update actualiza la lógica de la luciérnaga
func (f *Firefly) update(lanterns []*Lantern, dt float64) {
	// Actualizar fase de parpadeo
	f.updateBlinkPhase(dt)

	// Calcular fuerzas
	f.applyWandering()
	f.applyLanternAttraction(lanterns)
	f.applyAttractionPoint()
	f.applyWind()

	// Actualizar posición
	f.position = f.position.Add(f.velocity.Mul(dt))

	// Wrap around (efecto toroidal)
	f.position = utils.WrapAround(f.position, config.ScreenWidth, config.ScreenHeight)

	// Limitar velocidad
	if f.velocity.Magnitude() > config.FireflySpeed*2 {
		f.velocity = f.velocity.Normalize().Mul(config.FireflySpeed * 2)
	}
}

// updateBlinkPhase actualiza el brillo basado en un ciclo sinusoidal
func (f *Firefly) updateBlinkPhase(dt float64) {
	f.blinkPhase += dt / f.blinkCycleDur
	if f.blinkPhase > 1.0 {
		f.blinkPhase = 0.0
		// Variar ligeramente el ciclo para que no todas parpadeen igual
		f.blinkCycleDur = utils.RandomFloat(config.FireflyBlinkCycleMin, config.FireflyBlinkCycleMax)
	}

	// Función sinusoidal para el brillo (0.0 a 1.0)
	f.brightness = (math.Sin(f.blinkPhase*2*math.Pi) + 1) / 2
}

// applyWandering aplica movimiento aleatorio (comportamiento errático)
func (f *Firefly) applyWandering() {
	if utils.RandomFloat(0, 1) < 0.05 { // 5% de probabilidad cada frame
		randomForce := utils.RandomUnitVector().Mul(0.2)
		f.velocity = f.velocity.Add(randomForce)
	}
}

// applyLanternAttraction atrae la luciérnaga hacia faroles cercanos
func (f *Firefly) applyLanternAttraction(lanterns []*Lantern) {
	if lanterns == nil || len(lanterns) == 0 {
		return
	}

	for _, lantern := range lanterns {
		distance := utils.Distance(f.position, lantern.Position)

		// Solo atraer si está dentro del radio de influencia
		if distance < lantern.Radius && distance > 1 {
			direction := lantern.Position.Sub(f.position).Normalize()

			// Fuerza inversamente proporcional a la distancia
			strength := (lantern.Radius - distance) / lantern.Radius
			force := direction.Mul(config.LanternInfluenceForce * strength)

			f.velocity = f.velocity.Add(force)
		}
	}
}

// applyAttractionPoint atrae la luciérnaga hacia un punto específico (click del jugador)
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

// applyWind aplica la fuerza del viento
func (f *Firefly) applyWind() {
	if f.windForce == nil {
		return
	}

	// El viento afecta proporcionalmente con resistencia
	windEffect := f.windForce.Mul(config.FireflyWindResistance)
	f.velocity = f.velocity.Add(windEffect)
}

// SetAttractionPoint establece un punto de atracción para la luciérnaga
func (f *Firefly) SetAttractionPoint(point *utils.Vector2D) {
	f.attractionPoint = point
}

// SetWindForce establece la fuerza del viento
func (f *Firefly) SetWindForce(wind *utils.Vector2D) {
	f.windForce = wind
}

// GetID retorna el ID de la luciérnaga
func (f *Firefly) GetID() int {
	return f.id
}
