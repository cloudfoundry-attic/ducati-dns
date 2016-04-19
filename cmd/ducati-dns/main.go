package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-dns/cc_client"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/cloudfoundry-incubator/ducati-dns/uaa_client"
	"github.com/miekg/dns"
	"github.com/pivotal-cf-experimental/rainmaker"
	"github.com/pivotal-cf-experimental/warrant"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

type DNSConfig struct {
	Server            string
	DucatiSuffix      string
	DucatiAPI         string
	CCClientHost      string
	UAABaseURL        string
	UAAClientName     string
	UAASecret         string
	ListenAddress     string
	SkipSSLValidation bool
}

func (c DNSConfig) Validate() error {
	if c.Server == "" {
		return errors.New("missing required arg: server")
	}
	if c.DucatiSuffix == "" {
		return errors.New("missing required arg: ducatiSuffix")
	}
	if c.DucatiAPI == "" {
		return errors.New("missing required arg: ducatiAPI")
	}
	if c.CCClientHost == "" {
		return errors.New("missing required arg: ccAPI")
	}
	if c.UAAClientName == "" {
		return errors.New("missing required arg: uaaClientName")
	}
	if c.UAASecret == "" {
		return errors.New("missing required arg: uaaClientSecret")
	}
	if c.UAABaseURL == "" {
		return errors.New("missing required arg: uaaBaseURL")
	}

	return nil
}

func main() {
	var config DNSConfig

	flag.StringVar(&config.Server, "server", "", "Single DNS server to forward queries to")
	flag.StringVar(&config.DucatiSuffix, "ducatiSuffix", "", "suffix for lookups on the overlay network")
	flag.StringVar(&config.DucatiAPI, "ducatiAPI", "", "URL for the ducati API")
	flag.StringVar(&config.CCClientHost, "ccAPI", "", "URL for the cloud controller API")
	flag.StringVar(&config.UAABaseURL, "uaaBaseURL", "", "URL for the UAA API, e.g. https://uaa.example.com/")
	flag.StringVar(&config.UAAClientName, "uaaClientName", "", "client name for the UAA client")
	flag.StringVar(&config.UAASecret, "uaaClientSecret", "", "secret for the UAA client")
	flag.BoolVar(&config.SkipSSLValidation, "skipSSLValidation", false, "skip SSL validation for UAA")

	flag.StringVar(&config.ListenAddress, "listenAddress", "127.0.0.1:53", "Host and port to listen for queries on")
	flag.Parse()

	if err := config.Validate(); err != nil {
		log.Fatalf("validate: %s", err)
	}

	logger := lager.NewLogger("ducati-dns")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

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
	ducatiDaemonClient := client.New(config.DucatiAPI, http.DefaultClient)
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

	udpAddr, err := net.ResolveUDPAddr("udp", config.ListenAddress)
	if err != nil {
		log.Fatalf("invalid listen address %s: %s", config.ListenAddress, err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("listen: %s", err)
	}
	defer udpConn.Close()

	dnsRunner := &runner.Runner{
		DNSServer: &dns.Server{
			PacketConn: udpConn,
			Handler:    resolverMuxer,
		},
	}

	members := grouper.Members{
		{"dns_runner", dnsRunner},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	err = <-monitor.Wait()
	if err != nil {
		log.Fatalf("daemon terminated: %s", err)
	}
}
