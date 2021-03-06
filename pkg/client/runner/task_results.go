package runner

import (
	"fmt"
	"time"
)

type TaskResult struct {
	StartTime   uint64
	EndTime     uint64
	ProcessTime uint64
	Success     bool
	Category    string
	Err         string
}

type SerialTaskResult struct {
	SuccessNum  int
	FailureNum  int
	ProcessTime uint64
	SerialTime  uint64
	MaxTime     uint64
	MinTime     uint64
}

func (r *SerialTaskResult) Print() {
	fmt.Printf("tasks takes %d ms\n", r.SerialTime/uint64(time.Millisecond))
	fmt.Printf("process takes %d ms\n", r.ProcessTime/uint64(time.Millisecond))
	fmt.Printf("max process takes %d ms\n", r.MaxTime/uint64(time.Millisecond))
	fmt.Printf("min process takes %d ms\n", r.MinTime/uint64(time.Millisecond))
}
