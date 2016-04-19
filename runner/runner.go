package runner

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-dns/cc_client"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/uaa_client"
	"github.com/miekg/dns"
	"github.com/pivotal-cf-experimental/rainmaker"
	"github.com/pivotal-cf-experimental/warrant"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/dns_server.go --fake-name DNSServer . dnsServer
type dnsServer interface {
	ActivateAndServe() error
	Shutdown() error
}

type Config struct {
	Server            string
	DucatiSuffix      string
	DucatiAPI         string
	CCClientHost      string
	UAABaseURL        string
	UAAClientName     string
	UAASecret         string
	SkipSSLValidation bool
	Listener          net.PacketConn
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

	warrantClient := warrant.New(warrant.Config{
		Host:          config.UAABaseURL,
		SkipVerifySSL: config.SkipSSLValidation,
	})
	uaaClient := &uaa_client.Client{
		Service: warrantClient.Clients,
		User:    config.UAAClientName,
		Secret:  config.UAASecret,
	}
	rainmakerClient := rainmaker.NewClient(rainmaker.Config{
		Host: config.CCClientHost,
	})
	ccClient := cc_client.Client{
		Org:   rainmakerClient.Organizations,
		Space: rainmakerClient.Spaces,
		App:   rainmakerClient.Applications,
		UAA:   uaaClient,
	}
	ducatiDaemonClient := client.New(
		config.DucatiAPI,
		http.DefaultClient,
	)
	httpResolver := &resolver.HTTPResolver{
		Logger:       logger,
		Suffix:       config.DucatiSuffix,
		DaemonClient: ducatiDaemonClient,
		CCClient:     &ccClient,
	}
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
