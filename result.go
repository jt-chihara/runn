package runn

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

type RunResult struct {
	Desc string
	Path string
	Err  error
}

type runNResult struct {
	Total      atomic.Int64
	Success    atomic.Int64
	Failure    atomic.Int64
	Skipped    atomic.Int64
	RunResults sync.Map
}

func newRunResult(desc, path string) *RunResult {
	return &RunResult{
		Desc: desc,
		Path: path,
	}
}

func (r *runNResult) HasFailure() bool {
	return r.Failure.Load() > 0
}

func (r *runNResult) Out(out io.Writer) error {
	var ts, fs string
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	if r.Total.Load() == 1 {
		ts = fmt.Sprintf("%d scenario", r.Total.Load())
	} else {
		ts = fmt.Sprintf("%d scenarios", r.Total.Load())
	}
	ss := fmt.Sprintf("%d skipped", r.Skipped.Load())
	if r.Failure.Load() == 1 {
		fs = fmt.Sprintf("%d failure", r.Failure.Load())
	} else {
		fs = fmt.Sprintf("%d failures", r.Failure.Load())
	}
	if r.HasFailure() {
		if _, err := fmt.Fprintf(out, red("%s, %s, %s\n"), ts, ss, fs); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(out, green("%s, %s, %s\n"), ts, ss, fs); err != nil {
			return err
		}
	}
	return nil
}
