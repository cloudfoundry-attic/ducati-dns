package runner

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/dns_server.go --fake-name DNSServer . dnsServer
type dnsServer interface {
	ActivateAndServe() error
	Shutdown() error
}

type Config struct {
	resolver.Config
	Server   string
	Listener net.PacketConn
}

type Runner struct {
	DNSServer dnsServer
}

func New(logger lager.Logger, config Config) (*Runner, error) {
	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Server:    config.Server,
		Logger:    logger,
	}

	httpResolver := resolver.NewHTTPResolver(logger, config.Config)

	resolverMuxer := dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		if strings.HasSuffix(request.Question[0].Name, fmt.Sprintf("%s.", config.DucatiSuffix)) {
			httpResolver.ServeDNS(w, request)
		} else {
			forwardingResolver.ServeDNS(w, request)
		}
	})

	dnsRunner := &Runner{
		DNSServer: &dns.Server{
			PacketConn: config.Listener,
			Handler:    resolverMuxer,
		},
	}

	return dnsRunner, nil
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
