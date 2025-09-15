package workerpool

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransform(t *testing.T) {
	t.Run(
		"general flow",
		func(t *testing.T) {
			input, result := make(chan int), []string{}
			output := Transform(context.Background(), input, func(n int) string { return fmt.Sprint(n) }, 10)

			expected := []int{}
			go func() {
				defer close(input)
				for i := range 10_000 {
					expected = append(expected, i)
					input <- i
				}
			}()

			for v := range output {
				result = append(result, v)
			}

			actual := []int{}
			assert.LessOrEqual(t, len(result), 10_000)
			for _, v := range result {
				i, err := strconv.Atoi(v)
				assert.NoError(t, err)
				actual = append(actual, i)
			}
			assert.ElementsMatch(t, expected, actual)
		},
	)
	t.Run(
		"context cancel",
		func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			input, result := make(chan int), []string{}
			defer close(input)
			output := Transform(ctx, input, func(n int) string { return fmt.Sprint(n) }, 10)

			go func() { cancel() }()

			for v := range output {
				result = append(result, v)
			}

			assert.Empty(t, result)
		},
	)
	t.Run(
		"closed input channel",
		func(t *testing.T) {
			input, result := make(chan int), []string{}
			output := Transform(context.Background(), input, func(n int) string { return fmt.Sprint(n) }, 10)

			expected := []int{}
			go func() {
				defer close(input)
				for i := range 10_000 {
					if i == 5_000 {
						return
					}
					expected = append(expected, i)
					input <- i
				}
			}()

			for v := range output {
				result = append(result, v)
			}

			actual := []int{}
			assert.LessOrEqual(t, len(result), 10_000)
			for _, v := range result {
				i, err := strconv.Atoi(v)
				assert.NoError(t, err)
				actual = append(actual, i)
			}
			assert.ElementsMatch(t, expected, actual)
		},
	)
}
