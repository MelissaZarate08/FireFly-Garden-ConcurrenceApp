package manager

import (
	"context"
	"sync"
)

type Job struct {
	ID   int
	Task func() interface{}
}

type Result struct {
	JobID  int
	Output interface{}
	Error  error
}

type WorkerPool struct {
	workerCount int
	jobsCh      chan Job
	resultsCh   chan Result
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

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

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

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
			
			result := wp.processJob(job)
			
			select {
			case wp.resultsCh <- result:
			case <-wp.ctx.Done():
				return
			default:
			}
		}
	}
}

func (wp *WorkerPool) processJob(job Job) Result {
	result := Result{
		JobID: job.ID,
	}
	
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

func (wp *WorkerPool) GetJobsChannel() chan<- Job {
	return wp.jobsCh
}

func (wp *WorkerPool) GetResultsChannel() <-chan Result {
	return wp.resultsCh
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobsCh)
	wp.wg.Wait()
	close(wp.resultsCh)
}


func (wp *WorkerPool) WaitForCompletion() {
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
	
	for i := 0; i < barrierCount; i++ {
		<-barrierCh
	}
	
	close(barrierCh)
}