package network

import (
	"fmt"
	"math"
	"math/rand"
)

type task struct {
	workers  int
	duration float64
}

func (t *task) String() string {
	return fmt.Sprintf("task{workers: %d, duration: %.4f}", t.workers, t.duration)
}

type taskConfig struct {
	Rate               float64
	BaseWorkerCount    int
	WorkerScaleRange   float64
	BaseDuration       float64
	DurationScaleRange float64
}

func (tc *taskConfig) randomTask() task {
	randInRange := func(min, max float64) float64 {
		return rand.Float64()*(max-min) + min
	}
	workerScale := randInRange(1-tc.WorkerScaleRange, 1+tc.WorkerScaleRange)
	durationScale := randInRange(1-tc.DurationScaleRange, 1+tc.DurationScaleRange)
	return task{
		workers:  int(math.Floor(float64(tc.BaseWorkerCount) * workerScale)),
		duration: tc.BaseDuration * durationScale,
	}
}
