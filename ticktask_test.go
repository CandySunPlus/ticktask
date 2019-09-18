package ticktask

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskContext(t *testing.T) {
	start := time.Now()

	task := NewTickTask(func(ctx context.Context) error {
		if tickAt, exists := GetTickAt(ctx); exists {
			fmt.Printf("tick task run at %s \n", tickAt)
		}
		return errors.New("hello")
	}, Interval(time.Second*2), OnRetry(func(n uint, err error) {
		fmt.Printf("retry %d times: %s \n", n, err)
	}), Retry(3, time.Millisecond * 500))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)

	defer cancel()

	task.StartAndRun(ctx)

	task.Join()

	dur := time.Since(start)

	assert.True(t, dur >= time.Second*5, "task timeout after 5 second")
}
