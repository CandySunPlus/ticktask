package ticktask

import (
	"context"
	"fmt"
	"time"
)

type key string

var tickAt = key("tickAt")

type TaskFn func(ctx context.Context) error
type Option func(*TickTaskOptions)

type TickTaskOptions struct {
	interval   time.Duration
	retryDelay time.Duration
	attempts   int
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

func Retry(attempts int, retryDelay time.Duration) Option {
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

func (tt *tickTask) callTaskFn(ctx context.Context, attempts int) {
	if err := tt.taskFn(ctx); err != nil {
		if attempts--; attempts >= 0 {
			time.Sleep(tt.opts.retryDelay)
			tt.callTaskFn(ctx, attempts)
			return
		}
		fmt.Printf("retry task fault %d times: %s \n", tt.opts.attempts+1, err)
	}
}

func (tt *tickTask) StartAndRun(ctx context.Context) {
	ctx = context.WithValue(ctx, tickAt, time.Now())
	go tt.callTaskFn(ctx, tt.opts.attempts)
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
				go tt.callTaskFn(ctx, tt.opts.attempts)
			}
		}
	}()
}

func (tt *tickTask) Join() {
	<-tt.done
}
