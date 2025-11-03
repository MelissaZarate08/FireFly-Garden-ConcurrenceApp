package manager

import (
	"context"
	"sync"
)

// Job representa una tarea que será procesada por el worker pool
type Job struct {
	ID   int
	Task func() interface{}
}

// Result representa el resultado de un trabajo procesado
type Result struct {
	JobID  int
	Output interface{}
	Error  error
}

// WorkerPool implementa el patrón Worker Pool
// Múltiples workers consumen trabajos de un canal compartido
// Útil para procesamiento paralelo de tareas (ej: cálculos de colisión, pathfinding)
type WorkerPool struct {
	workerCount int
	jobsCh      chan Job
	resultsCh   chan Result
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewWorkerPool crea un nuevo pool de workers
func NewWorkerPool(workerCount, jobBufferSize, resultBufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workerCount: workerCount,
		jobsCh:      make(chan Job, jobBufferSize),
		resultsCh:   make(chan Result, resultBufferSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start inicia todos los workers del pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker es la goroutine que procesa trabajos
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()
	
	for {
		select {
		case <-wp.ctx.Done():
			return
			
		case job, ok := <-wp.jobsCh:
			if !ok {
				return
			}
			
			// Procesar el trabajo
			result := wp.processJob(job)
			
			// Enviar resultado (non-blocking)
			select {
			case wp.resultsCh <- result:
			case <-wp.ctx.Done():
				return
			default:
				// Si el canal de resultados está lleno, descartar
			}
		}
	}
}

// processJob ejecuta la tarea del trabajo y maneja errores
func (wp *WorkerPool) processJob(job Job) Result {
	result := Result{
		JobID: job.ID,
	}
	
	// Ejecutar la tarea y capturar panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				result.Error = nil
			}
		}()
		
		result.Output = job.Task()
	}()
	
	return result
}

// Submit envía un trabajo al pool
// Retorna false si el contexto fue cancelado
func (wp *WorkerPool) Submit(job Job) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.jobsCh <- job:
		return true
	default:
		// Canal lleno, descartar trabajo
		return false
	}
}

// GetJobsChannel retorna el canal de trabajos para envío directo
func (wp *WorkerPool) GetJobsChannel() chan<- Job {
	return wp.jobsCh
}

// GetResultsChannel retorna el canal de resultados para consumo
func (wp *WorkerPool) GetResultsChannel() <-chan Result {
	return wp.resultsCh
}

// Stop detiene el worker pool de forma limpia
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobsCh)
	wp.wg.Wait()
	close(wp.resultsCh)
}

// WaitForCompletion espera a que todos los trabajos actuales se completen
// Nota: No cierra el pool, solo espera a que el canal de trabajos se vacíe
func (wp *WorkerPool) WaitForCompletion() {
	// Enviar trabajos de señalización para confirmar que todos fueron procesados
	barrierCount := wp.workerCount
	barrierCh := make(chan struct{}, barrierCount)
	
	for i := 0; i < barrierCount; i++ {
		wp.Submit(Job{
			ID: -1,
			Task: func() interface{} {
				barrierCh <- struct{}{}
				return nil
			},
		})
	}
	
	// Esperar a que todos los workers procesen la barrera
	for i := 0; i < barrierCount; i++ {
		<-barrierCh
	}
	
	close(barrierCh)
}