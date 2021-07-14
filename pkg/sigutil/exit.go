package sigutil

import (
	"os"
	"os/signal"
	"syscall"
)

func ListenExit(fn func(os.Signal)) {
	go func() {
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)

		sig := <-sigchan
		if fn != nil {
			fn(sig)
		}
	}()
}
