package manager

import (
	"context"
	"sync"
	"time"

	"github.com/yourusername/firefly-garden/internal/config"
	"github.com/yourusername/firefly-garden/internal/core"
	"github.com/yourusername/firefly-garden/pkg/utils"
)

type Command struct {
	Type CommandType
	Data interface{}
}

type CommandType int

const (
	CommandSpawnFirefly CommandType = iota
	CommandSetAttraction
	CommandClearAttraction
	CommandUpdateWind
)

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

func NewFireflyManager() *FireflyManager {
	ctx, cancel := context.WithCancel(context.Background())

	aggregator := NewStateAggregator(config.StateChannelBuffer)
	wind := core.NewWind()

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

func (fm *FireflyManager) Start() {
	fm.aggregator.Start()

	fm.wg.Add(1)
	go fm.wind.Run(fm.ctx)
	go func() {
		<-fm.ctx.Done()
		fm.wg.Done()
	}()

	fm.workerPool.Start()

	fm.wg.Add(1)
	go fm.commandLoop()

	if config.AutoSpawnEnabled {
		fm.wg.Add(1)
		go fm.autoSpawner()
	}

	fm.spawnInitialFireflies()
}

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
		fm.wind.CycleDirection()
	}
}

func (fm *FireflyManager) autoSpawner() {
	defer fm.wg.Done()

	ticker := time.NewTicker(config.FireflySpawnInterval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-fm.ctx.Done():
			return

		case <-ticker.C:
			current := fm.GetFireflyCount()
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
				if missing > config.SpawnBurstCount*2 {
					x := utils.RandomFloat(0, config.ScreenWidth)
					y := utils.RandomFloat(0, config.ScreenHeight)
					fm.spawnFirefly(x, y)
				}
			} else {
				if utils.RandomFloat(0, 1) < 0.05 && fm.GetFireflyCount() < config.MaxFireflies {
					x := utils.RandomFloat(0, config.ScreenWidth)
					y := utils.RandomFloat(0, config.ScreenHeight)
					fm.spawnFirefly(x, y)
				}
			}
		}
	}
}

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


func (fm *FireflyManager) spawnInitialFireflies() {
	for i := 0; i < config.InitialFireflyCount; i++ {
		x := utils.RandomFloat(0, config.ScreenWidth)
		y := utils.RandomFloat(0, config.ScreenHeight)
		fm.spawnFirefly(x, y)
	}
}

func (fm *FireflyManager) spawnFirefly(x, y float64) {
	fm.firefliesMux.Lock()

	id := fm.nextID
	fm.nextID++

	firefly := core.NewFirefly(id, x, y)
	fm.fireflies[id] = firefly

	fm.firefliesMux.Unlock()

	firefly.SetWindForce(fm.wind.GetForcePointer())

	fm.attractionMux.RLock()
	if fm.attractionPt != nil {
		firefly.SetAttractionPoint(fm.attractionPt)
	}
	fm.attractionMux.RUnlock()

	lanterns := fm.getLanternsSnapshot()

	fm.wg.Add(1)
	go func(ff *core.Firefly, lns []*core.Lantern) {
		defer fm.wg.Done()
		ff.Run(fm.ctx, fm.aggregator.GetStateChannel(), lns, 1.0/float64(config.TargetFPS))
	}(firefly, lanterns)
}

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
		firefly.SetWindForce(fm.wind.GetForcePointer())
		if fm.attractionPt != nil {
			firefly.SetAttractionPoint(fm.attractionPt)
		}

		fm.fireflies[id] = firefly

		fm.wg.Add(1)
		go func(ff *core.Firefly) {
			defer fm.wg.Done()
			ff.Run(fm.ctx, fm.aggregator.GetStateChannel(), fm.getLanternsSnapshot(), 1.0/float64(config.TargetFPS))
		}(firefly)
	}
}

func (fm *FireflyManager) setAttractionPoint(point *utils.Vector2D) {
	fm.attractionMux.Lock()
	fm.attractionPt = point
	fm.attractionMux.Unlock()

	fm.firefliesMux.RLock()
	defer fm.firefliesMux.RUnlock()

	for _, firefly := range fm.fireflies {
		firefly.SetAttractionPoint(point)
	}
}

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

func (fm *FireflyManager) AddLantern(x, y float64) bool {
	fm.lanternsMux.Lock()
	defer fm.lanternsMux.Unlock()

	if len(fm.lanterns) >= config.MaxLanterns {
		return false
	}

	lantern := core.NewLantern(x, y)
	fm.lanterns = append(fm.lanterns, lantern)

	go fm.SpawnBurst(x, y, config.SpawnBurstCount)

	return true
}

func (fm *FireflyManager) RemoveLantern() {
	fm.lanternsMux.Lock()
	defer fm.lanternsMux.Unlock()

	if len(fm.lanterns) > 0 {
		fm.lanterns = fm.lanterns[:len(fm.lanterns)-1]
	}
}

func (fm *FireflyManager) UpdateLanterns(dt float64) {
	fm.lanternsMux.RLock()
	defer fm.lanternsMux.RUnlock()

	for _, lantern := range fm.lanterns {
		lantern.Update(dt)
	}
}

func (fm *FireflyManager) getLanternsSnapshot() []*core.Lantern {
	fm.lanternsMux.RLock()
	defer fm.lanternsMux.RUnlock()

	snapshot := make([]*core.Lantern, len(fm.lanterns))
	copy(snapshot, fm.lanterns)

	return snapshot
}

func (fm *FireflyManager) GetLanterns() []*core.Lantern {
	return fm.getLanternsSnapshot()
}

func (fm *FireflyManager) GetFireflyCount() int {
	return fm.aggregator.GetCount()
}

func (fm *FireflyManager) GetFireflyStates() []core.FireflyState {
	return fm.aggregator.GetSnapshot()
}

func (fm *FireflyManager) GetWind() *core.Wind {
	return fm.wind
}

func (fm *FireflyManager) GetCommandChannel() chan<- Command {
	return fm.commandCh
}

func (fm *FireflyManager) GetWorkerPool() *WorkerPool {
	return fm.workerPool
}

func (fm *FireflyManager) GetDroppedStates() uint64 {
	return core.GetDroppedStates()
}

func (fm *FireflyManager) Stop() {
	fm.cancel()

	fm.wg.Wait()

	fm.aggregator.Stop()
	fm.workerPool.Stop()

	close(fm.commandCh)
}
