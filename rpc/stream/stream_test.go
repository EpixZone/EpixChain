package stream

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamAdd(t *testing.T) {
	testCases := []struct {
		segmentSize, capacity int
	}{
		{128, 1280},
		{1024, 2048},
		{1024, 2148},
		{1024, 1024},
		{2048, 100},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("segmentSize=%d,capacity=%d", tc.segmentSize, tc.capacity)
		t.Run(name, func(t *testing.T) {
			stream := NewStream[int](tc.segmentSize, tc.capacity)

			amount := 100000
			for i := 0; i < amount; i++ {
				require.Equal(t, i+1, stream.Add(i))
			}

			all, _ := stream.ReadAllNonBlocking(0)
			maxSegments := (tc.capacity + tc.segmentSize - 1) / tc.segmentSize
			require.Equal(t, maxSegments*tc.segmentSize+amount%tc.segmentSize, len(all))
			require.Equal(t, 100000-1, all[len(all)-1])
			for i, n := range all[:len(all)-1] {
				require.Equal(t, n+1, all[i+1])
			}
		})
	}
}

func TestStreamReadNonBlocking(t *testing.T) {
	stream := NewStream[int](16, 31)

	for i := 0; i < 32; i++ {
		require.Equal(t, i+1, stream.Add(i))
	}

	items, offset := stream.ReadNonBlocking(0)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, items)
	require.Equal(t, 16, offset)
}

func TestStreamReadBlocking(t *testing.T) {
	const (
		total       = 32
		subscribers = 10
	)
	stream := NewStream[int](16, 31)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Each subscriber reads from the beginning (offset 0) so it deterministically
	// receives every published item regardless of goroutine scheduling: the ring
	// buffer retains all `total` items (they fit in 2 segments, so nothing is
	// pruned). This avoids the timing races of the old version, which relied on
	// fixed sleeps to (a) have every subscriber blocked before the publisher ran
	// and (b) have every subscriber drained before cancel — both of which are
	// unreliable under load / -race (a late subscriber caught only the last few
	// items). A completion barrier replaces the drain sleep.
	result := make([][]int, subscribers)
	var wg sync.WaitGroup
	done := make(chan struct{}, subscribers)

	for i := 0; i < subscribers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			offset := 0
			for len(result[i]) < total {
				items, next := stream.ReadBlocking(ctx, offset)
				if len(items) == 0 {
					return // context canceled
				}
				result[i] = append(result[i], items...)
				offset = next
			}
			done <- struct{}{}
		}(i)
	}

	// publisher
	for i := 0; i < total; i++ {
		require.Equal(t, i+1, stream.Add(i))
	}

	// wait until every subscriber has received all items (generous upper bound)
	deadline := time.After(30 * time.Second)
	for got := 0; got < subscribers; got++ {
		select {
		case <-done:
		case <-deadline:
			t.Fatalf("timed out waiting for subscribers: %d/%d completed", got, subscribers)
		}
	}

	cancel()
	wg.Wait()

	// check result: each subscriber saw the full, in-order sequence
	for i := 0; i < subscribers; i++ {
		require.Equal(t, total, len(result[i]))
		require.Equal(t, total-1, result[i][len(result[i])-1])
		for j := 0; j < len(result[i])-1; j++ {
			require.Equal(t, result[i][j]+1, result[i][j+1])
		}
	}
}

func TestStreamReadFromEnd(t *testing.T) {
	stream := NewStream[int](16, 31)

	items, offset := stream.ReadNonBlocking(-1)
	require.Empty(t, items)
	require.Equal(t, 0, offset)

	stream.Add(1)

	items, offset = stream.ReadNonBlocking(-1)
	require.Empty(t, items)
	require.Equal(t, 1, offset)

	stream.Add(2)

	items, offset = stream.ReadNonBlocking(offset)
	require.Equal(t, []int{2}, items)
	require.Equal(t, 2, offset)
}
