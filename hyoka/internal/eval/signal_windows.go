//go:build windows

package eval

import (
	"os"
	"os/signal"
)

// notifyShutdownSignals registers ch to receive os.Interrupt (Ctrl+C).
// Windows does not have SIGTERM; the closest equivalent is os.Interrupt
// which is delivered for both Ctrl+C and Ctrl+Break.
func notifyShutdownSignals(ch chan<- os.Signal) {
	signal.Notify(ch, os.Interrupt)
}
