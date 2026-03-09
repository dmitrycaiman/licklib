package ot05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestExtended(t *testing.T) {
	defer goleak.VerifyNone(t)
	type testCase struct{ taskCount, n, m int }

	for _, f := range []func(tasks []Task, n, m int) error{Run, RunWithSliceConcurrentAccess} {
		t.Run(
			"errors limit exceeded",
			func(t *testing.T) {
				cases := []*testCase{}
				for _, taskCount := range []int{0, 1, 10, 100} {
					for _, n := range []int{1, 10, 100} {
						cases = append(
							cases,
							&testCase{taskCount, n, taskCount},
							&testCase{taskCount, n, taskCount - 1},
							&testCase{taskCount, n, taskCount / 10},
							&testCase{taskCount, n, 0},
						)
					}
				}
				for _, c := range cases {
					tasks, finished := make([]Task, c.taskCount), new(atomic.Int64)
					for i := range c.taskCount {
						tasks[i] = func() error {
							time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
							finished.Add(1)
							return fmt.Errorf("%v", i)
						}
					}
					assert.ErrorIs(t, f(tasks, c.n, c.m), ErrErrorsLimitExceeded)
					assert.LessOrEqual(t, finished.Load(), int64(c.m+c.n))
				}
			},
		)
		t.Run(
			"no errors",
			func(t *testing.T) {
				cases := []*testCase{}
				for _, taskCount := range []int{10, 100, 1000} {
					for _, n := range []int{10, 100, 1000} {
						cases = append(cases, &testCase{taskCount, n, 1})
					}
				}
				for _, c := range cases {
					tasks, finished, elapsedSequential, elapsedParallel := make([]Task, c.taskCount), new(atomic.Int64), new(atomic.Int64), time.Duration(0)
					for i := range c.taskCount {
						tasks[i] = func() error {
							defer func(t time.Time) {
								elapsedSequential.Add(int64(time.Since(t)))
							}(time.Now())
							finished.Add(1)
							time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
							return nil
						}
					}
					assert.NoError(
						t,
						func() error {
							defer func(t time.Time) {
								elapsedParallel = time.Since(t)
							}(time.Now())
							return f(tasks, c.n, c.m)
						}(),
					)
					assert.Equal(t, finished.Load(), int64(c.taskCount))
					assert.Less(t, int64(elapsedParallel), elapsedSequential.Load())
				}
			},
		)
	}
}

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}
