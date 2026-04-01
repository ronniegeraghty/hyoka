//go:build !windows

package eval

import (
	"os"
	"os/signal"
	"syscall"
)

// notifyShutdownSignals registers ch to receive SIGINT and SIGTERM.
func notifyShutdownSignals(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}
