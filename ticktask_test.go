package ticktask

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskContext(t *testing.T) {
	start := time.Now()

	task := NewTickTask(func(ctx context.Context) error {
		return nil
	}, Interval(time.Second*1))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)

	defer cancel()

	task.Start(ctx)

	task.Join()

	dur := time.Since(start)

	assert.True(t, dur >= time.Second*5, "task timeout after 5 second")
}
