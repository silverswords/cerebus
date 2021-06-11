package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robertkrimen/otto"
)

// Task represents a generic task.
type Task interface {
	Do(context.Context) error

	BindScheduler(s *Scheduler) Task
	SetContext(context context.Context) Task
}

type CatchFunc func(err error)
type RetryTask interface {
	Task
	WithCatch(CatchFunc) Task
	WithRetry(times uint) Task
	WithTimeout(timeout time.Duration) Task
	WithCancelFunc(timeout time.Duration) (Task, context.CancelFunc)
}

type PriorityTask interface {
	Task
	WithPriority(int) Task
}

type CallbackFunc func(context.Context) error
type CallbackTask interface {
	Task
	AddStartCallback(f CallbackFunc) Task
	AddFinishedCallback(f CallbackFunc) Task
}

// TaskFunc is a wrapper for task function.
type TaskFunc func(context.Context) error

var _ Task = TaskFunc(func(context.Context) error { return nil })

// Do is the Task interface implementation for type TaskFunc.
func (t TaskFunc) Do(ctx context.Context) error {
	return t(ctx)
}

// WithCatch set the catch function for this task
func (t TaskFunc) WithCatch(f CatchFunc) Task {
	return &task{
		task:      t,
		catchFunc: f,
	}
}

// WithRetry set the retry times for this task
func (t TaskFunc) WithRetry(times uint) Task {
	task := &task{
		task: t,
	}

	return task.WithRetry(times)
}

// WithTimeout set the timeout for this task
func (t TaskFunc) WithTimeout(timeout time.Duration) Task {
	context, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	return &task{
		task:       t,
		ctx:        context,
		cancelFunc: cancelFunc,
	}
}

// WithCancelFunc returns the cancel function for this task
func (t TaskFunc) WithCancelFunc(timeout time.Duration) (Task, context.CancelFunc) {
	context, cancelFunc := context.WithCancel(context.Background())

	return &task{
		task:       t,
		ctx:        context,
		cancelFunc: cancelFunc,
	}, cancelFunc
}

// WithPriority set the priority for this task
func (t TaskFunc) WithPriority(priority int) Task {
	return &task{
		task:     t,
		priority: priority,
	}
}

// AddStartCallback add the start callback func to this task
func (t TaskFunc) AddStartCallback(f CallbackFunc) Task {
	return &task{
		task:          t,
		startCallBack: []CallbackFunc{f},
	}
}

// AddFinishedCallback add the finished callback func to this task
func (t TaskFunc) AddFinishedCallback(f CallbackFunc) Task {
	return &task{
		task:             t,
		finishedCallBack: []CallbackFunc{f},
	}
}

// BindScheduler bind the scheduler with this task, this shouldn't called by user
func (t TaskFunc) BindScheduler(s *Scheduler) Task {
	return &task{
		task: t,
		sche: s,
	}
}

// SetContext set the context for this task, the context will used when call the internal function
func (t TaskFunc) SetContext(ctx context.Context) Task {
	return &task{
		task: t,
		ctx:  ctx,
	}
}

// task is the implement for Task
type task struct {
	task       Task
	ctx        context.Context
	cancelFunc context.CancelFunc

	sche *Scheduler

	catchFunc  CatchFunc
	retryTimes uint
	timeout    time.Duration
	deadline   time.Time
	priority   int

	startCallBack    []CallbackFunc
	finishedCallBack []CallbackFunc
}

// NewTask return a task
func NewTask(f TaskFunc) Task {
	return &task{
		task: f,
	}
}

// Do is the Task interface implementation
func (t *task) Do(ctx context.Context) error {
	return t.task.Do(ctx)
}

// WithCatch set the catch function for this task
func (t *task) WithCatch(f CatchFunc) Task {
	t.catchFunc = f
	return t
}

// WithRetry set the retry times for this task
func (t *task) WithRetry(times uint) Task {
	counter, originTask := uint(0), t.task

	t.retryTimes = times
	t.task = TaskFunc(func(ctx context.Context) error {
		err := originTask.Do(ctx)
		if err == nil {
			return nil
		}

		log.Printf("[Task] error: %s", err)
		if counter < times {
			counter++
			log.Printf("[Task] Retry times: %d", counter)
			t.sche.queue.Add(t)
		}

		return nil
	})

	return t
}

// WithTimeout set the timeout for this task
func (t *task) WithTimeout(timeout time.Duration) Task {
	backgroundContext := context.Background()
	if t.ctx != nil {
		backgroundContext = t.ctx
	}
	t.deadline = time.Now().Add(timeout)
	context, cancelFunc := context.WithDeadline(backgroundContext, t.deadline)

	t.ctx = context
	t.cancelFunc = cancelFunc
	t.timeout = timeout

	return t
}

// WithCancelFunc returns the cancel function for this task
func (t *task) WithCancelFunc(timeout time.Duration) (Task, context.CancelFunc) {
	backgroundContext := context.Background()
	if t.ctx != nil {
		backgroundContext = t.ctx
	}

	context, cancelFunc := context.WithCancel(backgroundContext)
	t.ctx = context
	t.cancelFunc = cancelFunc
	return t, cancelFunc
}

// WithPriority set the priority for this task
func (t *task) WithPriority(priority int) Task {
	t.priority = priority
	return t
}

// AddStartCallback add the start callback func to this task
func (t *task) AddStartCallback(f CallbackFunc) Task {
	t.startCallBack = append(t.startCallBack, f)
	return t
}

// AddFinishedCallback add the finished callback func to this task
func (t *task) AddFinishedCallback(f CallbackFunc) Task {
	t.finishedCallBack = append(t.finishedCallBack, f)
	return t
}

// BindScheduler bind the scheduler with this task, this shouldn't called by user
func (t *task) BindScheduler(s *Scheduler) Task {
	t.sche = s
	return t
}

// SetContext set the context for this task, the context will used when call the internal function
func (t *task) SetContext(ctx context.Context) Task {
	if t.ctx != nil && t.ctx != ctx {
		log.Printf("[Warning] don't have the same context, use the lastest")
	}

	t.ctx = ctx
	return t
}

type JsTask struct {
	script string
	vm     *otto.Otto
}

func NewJsTask(script string) *JsTask {
	return &JsTask{
		script: script,
		vm:     otto.New(),
	}
}

func (t *JsTask) SetParam(name string, value interface{}) error {
	return t.vm.Set(name, value)
}

func (s *JsTask) Do(context.Context) error {
	_, err := s.vm.Run(s.script)
	if err != nil {
		return err
	}

	return nil
}

// WithCatch set the catch function for this task
func (t *JsTask) WithCatch(f CatchFunc) Task {
	return &task{
		task:      t,
		catchFunc: f,
	}
}

// WithRetry set the retry times for this task
func (t *JsTask) WithRetry(times uint) Task {
	task := &task{
		task: t,
	}

	return task.WithRetry(times)
}

// WithTimeout set the timeout for this task
func (t *JsTask) WithTimeout(timeout time.Duration) Task {
	context, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	return &task{
		task:       t,
		ctx:        context,
		cancelFunc: cancelFunc,
	}
}

// WithCancelFunc returns the cancel function for this task
func (t *JsTask) WithCancelFunc(timeout time.Duration) (Task, context.CancelFunc) {
	context, cancelFunc := context.WithCancel(context.Background())

	return &task{
		task:       t,
		ctx:        context,
		cancelFunc: cancelFunc,
	}, cancelFunc
}

// WithPriority set the priority for this task
func (t *JsTask) WithPriority(priority int) Task {
	return &task{
		task:     t,
		priority: priority,
	}
}

// AddStartCallback add the start callback func to this task
func (t *JsTask) AddStartCallback(f CallbackFunc) Task {
	return &task{
		task:          t,
		startCallBack: []CallbackFunc{f},
	}
}

// AddFinishedCallback add the finished callback func to this task
func (t *JsTask) AddFinishedCallback(f CallbackFunc) Task {
	return &task{
		task:             t,
		finishedCallBack: []CallbackFunc{f},
	}
}

// BindScheduler bind the scheduler with this task, this shouldn't called by user
func (t *JsTask) BindScheduler(s *Scheduler) Task {
	return &task{
		task: t,
		sche: s,
	}
}

// SetContext set the context for this task, the context will used when call the internal function
func (t *JsTask) SetContext(ctx context.Context) Task {
	return &task{
		task: t,
		ctx:  ctx,
	}
}
