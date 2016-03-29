package runner

import (
	"fmt"
	"os"
)

//go:generate counterfeiter -o ../fakes/dns_server.go --fake-name DNSServer . dnsServer
type dnsServer interface {
	ListenAndServe() error
	Shutdown() error
}

type Runner struct {
	DNSServer dnsServer
}

func (r *Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.DNSServer.ListenAndServe()
	}()

	close(ready)

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("listen and serve: %s", err)
			}
			return nil

		case <-signals:
			return r.DNSServer.Shutdown()
		}
	}
}
