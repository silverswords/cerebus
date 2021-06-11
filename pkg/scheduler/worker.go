package scheduler

import "errors"

type Worker interface {
	Work()
}

// Worker represents a working goroutine.
type goroutineWorker struct {
	sche   *Scheduler
	task   chan Task
	stopCh chan struct{}
}

func NewGoroutineWorker(s *Scheduler, stopCh chan struct{}) Worker {
	return &goroutineWorker{
		sche:   s,
		task:   make(chan Task),
		stopCh: stopCh,
	}
}

// StartWorker create a new worker.
func (s *Scheduler) startWorker(stopCh chan struct{}) {
	worker := NewGoroutineWorker(s, stopCh)

	go worker.Work()
}

// Worker's main loop.
func (w *goroutineWorker) Work() {
	w.sche.workers <- w.task

	for {
		select {
		case t := <-w.task:
			realTask := t.(*task)

			defer func() {
				if r := recover(); r != nil {
					realTask.sche.queue.Done(t)
					return
				}
			}()

			select {
			case <-realTask.ctx.Done():
				realTask.sche.queue.Done(t)
				realTask.cancelFunc()
				if realTask.catchFunc != nil {
					realTask.catchFunc(errors.New("Task cancel"))
					break
				}
				return
			default:
			}

			for _, f := range realTask.startCallBack {
				if err := f(realTask.ctx); err != nil {
					if realTask.catchFunc != nil {
						realTask.catchFunc(err)
						break
					}
				}
			}

			if err := realTask.Do(realTask.ctx); err != nil {
				if realTask.catchFunc != nil {
					realTask.catchFunc(err)
					break
				}
			}

			realTask.sche.queue.Done(t)
			for _, f := range realTask.finishedCallBack {
				if err := f(realTask.ctx); err != nil {
					if realTask.catchFunc != nil {
						realTask.catchFunc(err)
						break
					}
				}

			}

			w.sche.workers <- w.task
		case <-w.stopCh:
			close(w.task)
			return
		}
	}
}
