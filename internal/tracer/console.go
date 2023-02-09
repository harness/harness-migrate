// Copyright 2023 Harness Inc. All rights reserved.

package tracer

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type console struct {
	bar  *progressbar.ProgressBar
	time time.Time
	done chan (bool)
	once sync.Once
}

// New returns a tracer that outputs the outputs the
// progress to the terminal.
func New() *console {
	return &console{
		done: make(chan (bool)),
		bar: progressbar.NewOptions64(-1,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
		),
	}
}

// Start starts the trace routine.
func (c *console) Start(format string, args ...interface{}) {
	c.once.Do(func() {
		go c.start()
	})
	c.time = time.Now()
	c.bar.Describe(fmt.Sprintf(format, args...))
}

// Stop stops the trace routine.
func (c *console) Stop(format string, args ...interface{}) {
	// this code implements an artificial delay to
	// prevent the progress bars from appearing and
	// disapparing too quickly.
	if time.Now().Sub(c.time) < (time.Second) {
		time.Sleep(time.Second)
	}

	c.bar.Clear()
	fmt.Printf(format, args...)
	fmt.Println("")
}

// Close stops the progress bar.
func (c *console) Close() {
	c.done <- true
	c.bar.Exit()
}

// start starts the progress bar.
func (c *console) start() {
	go func() {
		for {
			select {
			case <-c.done:
				return
			default:
				c.bar.Add(1)
				time.Sleep(time.Millisecond * 65)
			}
		}
	}()
}
