package ticktask

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskContext(t *testing.T) {
	start := time.Now()

	task := NewTickTask(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("task stop with context done")
				return nil
			default:
				if tickAt, exists := GetTickAt(ctx); exists {
					fmt.Printf("tick task run at %s \n", tickAt)
				}
				time.Sleep(time.Second * 2)
			}
		}
	}, Interval(time.Second*2))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)

	defer cancel()

	task.Start(ctx)

	task.Join()

	dur := time.Since(start)

	assert.True(t, dur >= time.Second*5, "task timeout after 5 second")
}
