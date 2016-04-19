package runner

import (
	"fmt"
	"os"
)

//go:generate counterfeiter -o ../fakes/dns_server.go --fake-name DNSServer . dnsServer
type dnsServer interface {
	ActivateAndServe() error
	Shutdown() error
}

type Runner struct {
	DNSServer dnsServer
}

func (r *Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- r.DNSServer.ActivateAndServe()
	}()

	close(ready)

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("activate and serve: %s", err)
			}
			return nil

		case <-signals:
			return r.DNSServer.Shutdown()
		}
	}
}
