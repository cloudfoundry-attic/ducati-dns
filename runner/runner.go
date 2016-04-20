package runner

import (
	"fmt"
	"net"
	"os"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/dns_server.go --fake-name DNSServer . dnsServer
type dnsServer interface {
	ActivateAndServe() error
	Shutdown() error
}

type Runner struct {
	DNSServer dnsServer
}

func New(
	logger lager.Logger,
	config resolver.Config,
	externalDNSServer string,
	listener net.PacketConn,
) *Runner {
	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Server:    externalDNSServer,
		Logger:    logger,
	}

	httpResolver := resolver.NewHTTPResolver(logger, config)

	resolverMuxer := &resolver.Muxer{
		Logger:               logger,
		Suffix:               config.DucatiSuffix,
		SuffixPresentHandler: httpResolver,
		DefaultHandler:       forwardingResolver,
	}

	dnsRunner := &Runner{
		DNSServer: &dns.Server{
			PacketConn: listener,
			Handler:    resolverMuxer,
		},
	}

	return dnsRunner
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
