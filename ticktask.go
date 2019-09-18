package ticktask

import (
	"context"
	"fmt"
	"time"
)

type key string

var tickAt = key("tickAt")

type TaskFn func(context.Context) error
type Option func(*TickTaskOptions)
type OnRetryFn func(uint, error)

type TickTaskOptions struct {
	interval   time.Duration
	retryDelay time.Duration
	onRetry    OnRetryFn
	attempts   uint
}

type tickTask struct {
	opts   *TickTaskOptions
	taskFn TaskFn
	done   chan bool
}

func GetTickAt(ctx context.Context) (time.Time, bool) {
	tickAt, exists := ctx.Value(tickAt).(time.Time)
	return tickAt, exists
}

func Interval(interval time.Duration) Option {
	return func(opts *TickTaskOptions) {
		opts.interval = interval
	}
}

func OnRetry(onRetryFn OnRetryFn) Option {
	return func(opts *TickTaskOptions) {
		opts.onRetry = onRetryFn
	}
}

func Retry(attempts uint, retryDelay time.Duration) Option {
	return func(opts *TickTaskOptions) {
		opts.retryDelay = retryDelay
		opts.attempts = attempts
	}
}

func NewTickTask(taskFn TaskFn, opts ...Option) *tickTask {
	options := &TickTaskOptions{
		interval:   time.Second * 5,
		retryDelay: time.Second * 1,
		attempts:   0,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &tickTask{
		opts:   options,
		taskFn: taskFn,
		done:   make(chan bool),
	}
}

func (tt *tickTask) callTaskFn(ctx context.Context) {
	var n uint = 1
	for n < tt.opts.attempts+1 {
		if err := tt.taskFn(ctx); err != nil {
			tt.opts.onRetry(n, err)
			if n == tt.opts.attempts {
				return
			}
			time.Sleep(tt.opts.retryDelay)
		} else {
			return
		}
		n++
	}
}

func (tt *tickTask) StartAndRun(ctx context.Context) {
	ctx = context.WithValue(ctx, tickAt, time.Now())
	go tt.callTaskFn(ctx)
	tt.Start(ctx)
}

func (tt *tickTask) Start(ctx context.Context) {
	ticker := time.NewTicker(tt.opts.interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context has been done, ticker stopped")
				tt.done <- true
				return
			case t := <-ticker.C:
				ctx = context.WithValue(ctx, tickAt, t)
				go tt.callTaskFn(ctx)
			}
		}
	}()
}

func (tt *tickTask) Join() {
	<-tt.done
}
