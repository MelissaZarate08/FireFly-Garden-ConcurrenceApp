package manager

import (
	"context"
	"sync"

	"github.com/yourusername/firefly-garden/internal/core"
)

type StateAggregator struct {
	states     map[int]core.FireflyState
	statesMux  sync.RWMutex
	stateCh    chan core.FireflyState
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewStateAggregator(bufferSize int) *StateAggregator {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StateAggregator{
		states:  make(map[int]core.FireflyState),
		stateCh: make(chan core.FireflyState, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (sa *StateAggregator) Start() {
	sa.wg.Add(1)
	go sa.aggregateLoop()
}

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

func (sa *StateAggregator) updateState(state core.FireflyState) {
	sa.statesMux.Lock()
	defer sa.statesMux.Unlock()
	
	if state.IsAlive {
		sa.states[state.ID] = state
	} else {
		delete(sa.states, state.ID)
	}
}

func (sa *StateAggregator) GetSnapshot() []core.FireflyState {
	sa.statesMux.RLock()
	defer sa.statesMux.RUnlock()
	
	snapshot := make([]core.FireflyState, 0, len(sa.states))
	
	for _, state := range sa.states {
		snapshot = append(snapshot, state)
	}
	
	return snapshot
}

func (sa *StateAggregator) GetCount() int {
	sa.statesMux.RLock()
	defer sa.statesMux.RUnlock()
	
	return len(sa.states)
}

func (sa *StateAggregator) GetStateChannel() chan<- core.FireflyState {
	return sa.stateCh
}

func (sa *StateAggregator) Clear() {
	sa.statesMux.Lock()
	defer sa.statesMux.Unlock()
	
	sa.states = make(map[int]core.FireflyState)
}

func (sa *StateAggregator) Stop() {
	sa.cancel()
	sa.wg.Wait()
	close(sa.stateCh)
}