package spinner

import (
	"fmt"
	"time"
)

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Start prints a spinning animation with msg until the returned stop func is called.
// stop() blocks until the goroutine has fully exited and the line is cleared.
func Start(msg string) func() {
	done := make(chan struct{})
	exited := make(chan struct{})

	go func() {
		defer close(exited)
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r\033[K") // clear spinner line
				return
			default:
				fmt.Printf("\r%s %s", frames[i%len(frames)], msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	return func() {
		close(done)
		<-exited // wait until goroutine has cleared the line
	}
}
