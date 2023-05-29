package gfsm

import "testing"

func BenchmarkSliceAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sliceTest()
	}
}

func BenchmarkMapAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mapTest()
	}
}

func sliceTest() bool {
	transitions := map[StartStopSM]struct{}{
		Stop:       {},
		Start:      {},
		InProgress: {},
	}
	_, found := transitions[Stop]
	return found
}

func mapTest() bool {
	transitions := []StartStopSM{
		Start, Stop, InProgress,
	}
	found := false
	for _, transition := range transitions {
		if Stop == transition {
			found = true
			break
		}
	}
	return found
}
