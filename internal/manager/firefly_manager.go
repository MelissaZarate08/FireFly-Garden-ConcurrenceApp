package manager

import (
	"context"
	"sync"
	"time"

	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/core"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

// Command representa comandos que se pueden enviar al manager
type Command struct {
	Type CommandType
	Data interface{}
}

// CommandType define los tipos de comandos disponibles
type CommandType int

const (
	CommandSpawnFirefly CommandType = iota
	CommandSetAttraction
	CommandClearAttraction
	CommandUpdateWind
)

// FireflyManager es el gestor principal que implementa el patrón Productor
// Crea y coordina todas las luciérnagas (goroutines)
type FireflyManager struct {
	fireflies      map[int]*core.Firefly
	firefliesMux   sync.RWMutex
	nextID         int
	aggregator     *StateAggregator
	wind           *core.Wind
	lanterns       []*core.Lantern
	lanternsMux    sync.RWMutex
	commandCh      chan Command
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	workerPool     *WorkerPool
	attractionPt   *utils.Vector2D
	attractionMux  sync.RWMutex
}

// NewFireflyManager crea un nuevo gestor de luciérnagas
func NewFireflyManager() *FireflyManager {
	ctx, cancel := context.WithCancel(context.Background())

	aggregator := NewStateAggregator(config.StateChannelBuffer)
	wind := core.NewWind()

	// Worker pool con 4 workers para procesamiento paralelo
	workerPool := NewWorkerPool(4, 100, 100)

	return &FireflyManager{
		fireflies:  make(map[int]*core.Firefly),
		nextID:     1,
		aggregator: aggregator,
		wind:       wind,
		lanterns:   make([]*core.Lantern, 0, config.MaxLanterns),
		commandCh:  make(chan Command, config.CommandChannelBuffer),
		ctx:        ctx,
		cancel:     cancel,
		workerPool: workerPool,
	}
}

// Start inicia el manager y sus componentes
func (fm *FireflyManager) Start() {
	// Iniciar agregador de estados (Fan-in)
	fm.aggregator.Start()

	// Iniciar sistema de viento
	fm.wg.Add(1)
	go fm.wind.Run(fm.ctx)
	go func() {
		<-fm.ctx.Done()
		fm.wg.Done()
	}()

	// Iniciar worker pool
	fm.workerPool.Start()

	// Iniciar loop de comandos
	fm.wg.Add(1)
	go fm.commandLoop()

	// Iniciar spawner automático (si está habilitado)
	if config.AutoSpawnEnabled {
		fm.wg.Add(1)
		go fm.autoSpawner()
	}

	// Spawner inicial
	fm.spawnInitialFireflies()
}

// commandLoop procesa comandos recibidos
func (fm *FireflyManager) commandLoop() {
	defer fm.wg.Done()

	for {
		select {
		case <-fm.ctx.Done():
			return

		case cmd := <-fm.commandCh:
			fm.processCommand(cmd)
		}
	}
}

// processCommand ejecuta un comando
func (fm *FireflyManager) processCommand(cmd Command) {
	switch cmd.Type {
	case CommandSpawnFirefly:
		pos, ok := cmd.Data.(utils.Vector2D)
		if ok {
			fm.spawnFirefly(pos.X, pos.Y)
		}

	case CommandSetAttraction:
		pos, ok := cmd.Data.(utils.Vector2D)
		if ok {
			fm.setAttractionPoint(&pos)
		}

	case CommandClearAttraction:
		fm.clearAttractionPoint()

	case CommandUpdateWind:
		// El viento ya se actualiza automáticamente
		fm.wind.CycleDirection()
	}
}

// autoSpawner genera luciérnagas automáticamente (Patrón Productor)
// Ahora rellena hasta ObjectiveCount en ráfagas, con mayor frecuencia si se está lejos del objetivo.
func (fm *FireflyManager) autoSpawner() {
	defer fm.wg.Done()

	// Ticker más dinámico (spawn más frecuente si falta alcanzar objetivo)
	ticker := time.NewTicker(config.FireflySpawnInterval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return

		case <-ticker.C:
			current := fm.GetFireflyCount()
			// Si estamos por debajo del objetivo, spawn en ráfaga aleatoria
			if current < config.ObjectiveCount {
				missing := config.ObjectiveCount - current
				toSpawn := config.SpawnBurstCount
				if missing < toSpawn {
					toSpawn = missing
				}
				for i := 0; i < toSpawn; i++ {
					x := utils.RandomFloat(0, config.ScreenWidth)
					y := utils.RandomFloat(0, config.ScreenHeight)
					fm.spawnFirefly(x, y)
				}
				// si estamos muy por debajo, spawn un poco más
				if missing > config.SpawnBurstCount*2 {
					x := utils.RandomFloat(0, config.ScreenWidth)
					y := utils.RandomFloat(0, config.ScreenHeight)
					fm.spawnFirefly(x, y)
				}
			} else {
				// Ocasionalmente crear 1 para mantener dinámica
				if utils.RandomFloat(0, 1) < 0.05 && fm.GetFireflyCount() < config.MaxFireflies {
					x := utils.RandomFloat(0, config.ScreenWidth)
					y := utils.RandomFloat(0, config.ScreenHeight)
					fm.spawnFirefly(x, y)
				}
			}
		}
	}
}

// autoSpawner simple fallback (no usado si config.AutoSpawnEnabled=false)
func (fm *FireflyManager) autoSpawnerSimple() {
	defer fm.wg.Done()
	ticker := time.NewTicker(config.FireflySpawnInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return
		case <-ticker.C:
			if fm.GetFireflyCount() < config.MaxFireflies {
				x := utils.RandomFloat(0, config.ScreenWidth)
				y := utils.RandomFloat(0, config.ScreenHeight)
				fm.spawnFirefly(x, y)
			}
		}
	}
}

// autoSpawner termina

// spawnInitialFireflies crea las luciérnagas iniciales
func (fm *FireflyManager) spawnInitialFireflies() {
	for i := 0; i < config.InitialFireflyCount; i++ {
		x := utils.RandomFloat(0, config.ScreenWidth)
		y := utils.RandomFloat(0, config.ScreenHeight)
		fm.spawnFirefly(x, y)
	}
}

// spawnFirefly crea una nueva luciérnaga y lanza su goroutine
func (fm *FireflyManager) spawnFirefly(x, y float64) {
	fm.firefliesMux.Lock()

	id := fm.nextID
	fm.nextID++

	firefly := core.NewFirefly(id, x, y)
	fm.fireflies[id] = firefly

	fm.firefliesMux.Unlock()

	// Configurar referencias compartidas
	firefly.SetWindForce(fm.wind.GetForcePointer())

	fm.attractionMux.RLock()
	if fm.attractionPt != nil {
		firefly.SetAttractionPoint(fm.attractionPt)
	}
	fm.attractionMux.RUnlock()

	// Obtener snapshot de faroles para esta luciérnaga
	lanterns := fm.getLanternsSnapshot()

	// Lanzar goroutine de la luciérnaga (Fan-out)
	fm.wg.Add(1)
	go func(ff *core.Firefly, lns []*core.Lantern) {
		defer fm.wg.Done()
		ff.Run(fm.ctx, fm.aggregator.GetStateChannel(), lns, 1.0/float64(config.TargetFPS))
	}(firefly, lanterns)
}

// SpawnBurst crea varias luciérnagas alrededor de una posición (útil para feedback inmediato)
func (fm *FireflyManager) SpawnBurst(x, y float64, count int) {
	fm.firefliesMux.Lock()
	defer fm.firefliesMux.Unlock()

	for i := 0; i < count; i++ {
		if len(fm.fireflies) >= config.MaxFireflies {
			return
		}
		dx := utils.RandomFloat(-40, 40)
		dy := utils.RandomFloat(-40, 40)
		id := fm.nextID
		fm.nextID++

		firefly := core.NewFirefly(id, x+dx, y+dy)
		// configurar referencias compartidas
		firefly.SetWindForce(fm.wind.GetForcePointer())
		if fm.attractionPt != nil {
			firefly.SetAttractionPoint(fm.attractionPt)
		}

		fm.fireflies[id] = firefly

		// lanzar goroutine para la nueva luciérnaga
		fm.wg.Add(1)
		go func(ff *core.Firefly) {
			defer fm.wg.Done()
			ff.Run(fm.ctx, fm.aggregator.GetStateChannel(), fm.getLanternsSnapshot(), 1.0/float64(config.TargetFPS))
		}(firefly)
	}
}

// setAttractionPoint establece un punto de atracción global
func (fm *FireflyManager) setAttractionPoint(point *utils.Vector2D) {
	fm.attractionMux.Lock()
	fm.attractionPt = point
	fm.attractionMux.Unlock()

	// Actualizar todas las luciérnagas existentes
	fm.firefliesMux.RLock()
	defer fm.firefliesMux.RUnlock()

	for _, firefly := range fm.fireflies {
		firefly.SetAttractionPoint(point)
	}
}

// clearAttractionPoint elimina el punto de atracción
func (fm *FireflyManager) clearAttractionPoint() {
	fm.attractionMux.Lock()
	fm.attractionPt = nil
	fm.attractionMux.Unlock()

	fm.firefliesMux.RLock()
	defer fm.firefliesMux.RUnlock()

	for _, firefly := range fm.fireflies {
		firefly.SetAttractionPoint(nil)
	}
}

// AddLantern agrega un nuevo farol (genera ráfaga como feedback)
func (fm *FireflyManager) AddLantern(x, y float64) bool {
	fm.lanternsMux.Lock()
	defer fm.lanternsMux.Unlock()

	if len(fm.lanterns) >= config.MaxLanterns {
		return false
	}

	lantern := core.NewLantern(x, y)
	fm.lanterns = append(fm.lanterns, lantern)

	// Feedback inmediato: generar una pequeña ráfaga alrededor del farol
	go fm.SpawnBurst(x, y, config.SpawnBurstCount)

	return true
}

// RemoveLantern elimina el último farol añadido
func (fm *FireflyManager) RemoveLantern() {
	fm.lanternsMux.Lock()
	defer fm.lanternsMux.Unlock()

	if len(fm.lanterns) > 0 {
		fm.lanterns = fm.lanterns[:len(fm.lanterns)-1]
	}
}

// UpdateLanterns actualiza todos los faroles
func (fm *FireflyManager) UpdateLanterns(dt float64) {
	fm.lanternsMux.RLock()
	defer fm.lanternsMux.RUnlock()

	for _, lantern := range fm.lanterns {
		lantern.Update(dt)
	}
}

// getLanternsSnapshot retorna una copia thread-safe de los faroles
func (fm *FireflyManager) getLanternsSnapshot() []*core.Lantern {
	fm.lanternsMux.RLock()
	defer fm.lanternsMux.RUnlock()

	snapshot := make([]*core.Lantern, len(fm.lanterns))
	copy(snapshot, fm.lanterns)

	return snapshot
}

// GetLanterns retorna los faroles para renderizado
func (fm *FireflyManager) GetLanterns() []*core.Lantern {
	return fm.getLanternsSnapshot()
}

// GetFireflyCount retorna el número actual de luciérnagas
func (fm *FireflyManager) GetFireflyCount() int {
	return fm.aggregator.GetCount()
}

// GetFireflyStates retorna un snapshot de todos los estados
func (fm *FireflyManager) GetFireflyStates() []core.FireflyState {
	return fm.aggregator.GetSnapshot()
}

// GetWind retorna el sistema de viento
func (fm *FireflyManager) GetWind() *core.Wind {
	return fm.wind
}

// GetCommandChannel retorna el canal de comandos
func (fm *FireflyManager) GetCommandChannel() chan<- Command {
	return fm.commandCh
}

// GetWorkerPool retorna el worker pool para procesamiento paralelo
func (fm *FireflyManager) GetWorkerPool() *WorkerPool {
	return fm.workerPool
}

// GetDroppedStates retorna la cantidad de estados descartados (por canales llenos)
func (fm *FireflyManager) GetDroppedStates() uint64 {
	return core.GetDroppedStates()
}

// Stop detiene el manager y todas sus goroutines de forma limpia
func (fm *FireflyManager) Stop() {
	// Cancelar contexto (señal de parada para todas las goroutines)
	fm.cancel()

	// Esperar a que todas las goroutines terminen
	fm.wg.Wait()

	// Detener componentes
	fm.aggregator.Stop()
	fm.workerPool.Stop()

	// Cerrar canal de comandos
	close(fm.commandCh)
}
