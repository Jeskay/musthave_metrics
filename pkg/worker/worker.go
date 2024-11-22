package worker

import "sync"

type WorkerPool[T any] struct {
	count int
}

func NewWorkerPool[T any](workerCount int) *WorkerPool[T] {
	return &WorkerPool[T]{workerCount}
}

func (wp *WorkerPool[T]) Run(jobs <-chan T, handler func(T)) {
	var wg sync.WaitGroup
	for range wp.count {
		wg.Add(1)
		go func() {
			for j := range jobs {
				handler(j)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
