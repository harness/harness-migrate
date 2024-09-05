// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracer

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

type console struct {
	bar      *progressbar.ProgressBar
	time     time.Time
	done     chan (bool)
	once     sync.Once
	logLevel LogLevel
}

// Log logs a message to the console.
func (c *console) Log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Println("")
}

// LogError logs an error message to the console.
func (c *console) LogError(format string, args ...interface{}) {
	// make sure only print the error message if passed
	modifiedFormat := strings.ReplaceAll(format, "%w", "%v")
	fmt.Println()
	color.Red(modifiedFormat, args...)
}

// New returns a tracer that outputs
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
	// make sure only print the error message if passed
	modifiedFormat := strings.ReplaceAll(format, "%w", "%v")
	fmt.Printf(modifiedFormat, args...)
	fmt.Println("")
}

func (c *console) Debug() Tracer {
	if c.logLevel == LogLevelDebug {
		return c
	}
	return none{}
}

// Close stops the progress bar.
func (c *console) Close() {
	close(c.done)
	c.bar.Exit()
}

func (c *console) WithLevel(level LogLevel) {
	c.logLevel = level
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
