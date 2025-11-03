package manager

import (
	"context"
	"sync"

	"github.com/yourusername/firefly-garden/internal/core"
)

// StateAggregator implementa el patrón Fan-in
// Recolecta estados de múltiples goroutines de luciérnagas
// y proporciona un snapshot thread-safe para el renderer
type StateAggregator struct {
	states     map[int]core.FireflyState
	statesMux  sync.RWMutex
	stateCh    chan core.FireflyState
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewStateAggregator crea un nuevo agregador de estados
func NewStateAggregator(bufferSize int) *StateAggregator {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StateAggregator{
		states:  make(map[int]core.FireflyState),
		stateCh: make(chan core.FireflyState, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start inicia la goroutine que agrega estados
// Esta es la parte "Fan-in": múltiples productores (luciérnagas)
// envían a un solo consumidor (este agregador)
func (sa *StateAggregator) Start() {
	sa.wg.Add(1)
	go sa.aggregateLoop()
}

// aggregateLoop es el loop principal que consume estados del canal
func (sa *StateAggregator) aggregateLoop() {
	defer sa.wg.Done()
	
	for {
		select {
		case <-sa.ctx.Done():
			return
			
		case state := <-sa.stateCh:
			sa.updateState(state)
		}
	}
}

// updateState actualiza el mapa de estados de forma thread-safe
func (sa *StateAggregator) updateState(state core.FireflyState) {
	sa.statesMux.Lock()
	defer sa.statesMux.Unlock()
	
	if state.IsAlive {
		// Actualizar estado existente o agregar nuevo
		sa.states[state.ID] = state
	} else {
		// Eliminar luciérnaga muerta
		delete(sa.states, state.ID)
	}
}

// GetSnapshot retorna una copia inmutable de todos los estados
// Esta función es thread-safe y puede ser llamada desde el renderer
func (sa *StateAggregator) GetSnapshot() []core.FireflyState {
	sa.statesMux.RLock()
	defer sa.statesMux.RUnlock()
	
	// Crear slice con capacidad conocida para eficiencia
	snapshot := make([]core.FireflyState, 0, len(sa.states))
	
	for _, state := range sa.states {
		snapshot = append(snapshot, state)
	}
	
	return snapshot
}

// GetCount retorna el número actual de luciérnagas vivas
func (sa *StateAggregator) GetCount() int {
	sa.statesMux.RLock()
	defer sa.statesMux.RUnlock()
	
	return len(sa.states)
}

// GetStateChannel retorna el canal para que las luciérnagas publiquen estados
func (sa *StateAggregator) GetStateChannel() chan<- core.FireflyState {
	return sa.stateCh
}

// Clear limpia todos los estados almacenados
func (sa *StateAggregator) Clear() {
	sa.statesMux.Lock()
	defer sa.statesMux.Unlock()
	
	sa.states = make(map[int]core.FireflyState)
}

// Stop detiene el agregador de forma limpia
func (sa *StateAggregator) Stop() {
	sa.cancel()
	sa.wg.Wait()
	close(sa.stateCh)
}