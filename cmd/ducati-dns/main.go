package main

import (
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

func main() {
	var (
		server            string
		ducatiSuffix      string
		ducatiAPI         string
		ccClientHost      string
		uaaBaseURL        string
		uaaClientName     string
		uaaSecret         string
		skipSSLValidation bool
	)

	flag.StringVar(&server, "server", "", "Single DNS server to forward queries to")
	flag.StringVar(&ducatiSuffix, "ducatiSuffix", "", "suffix for lookups on the overlay network")
	flag.StringVar(&ducatiAPI, "ducatiAPI", "", "URL for the ducati API")
	flag.StringVar(&ccClientHost, "ccAPI", "", "URL for the cloud controller API")
	flag.StringVar(&uaaBaseURL, "uaaBaseURL", "", "URL for the UAA API, e.g. https://uaa.example.com/")
	flag.StringVar(&uaaClientName, "uaaClientName", "", "client name for the UAA client")
	flag.StringVar(&uaaSecret, "uaaClientSecret", "", "secret for the UAA client")
	flag.BoolVar(&skipSSLValidation, "skipSSLValidation", false, "skip SSL validation for UAA")

	var listenAddress string
	flag.StringVar(&listenAddress, "listenAddress", "127.0.0.1:53", "Host and port to listen for queries on")
	flag.Parse()

	if server == "" {
		log.Fatalf("missing required arg: server")
	}
	if ducatiSuffix == "" {
		log.Fatalf("missing required arg: ducatiSuffix")
	}
	if ducatiAPI == "" {
		log.Fatalf("missing required arg: ducatiAPI")
	}
	if uaaClientName == "" {
		log.Fatalf("missing required arg: uaaClientName")
	}
	if uaaSecret == "" {
		log.Fatalf("missing required arg: uaaClientSecret")
	}
	if uaaBaseURL == "" {
		log.Fatalf("missing required arg: uaaBaseURL")
	}

	logger := lager.NewLogger("ducati-dns")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Server:    server,
		Logger:    logger,
	}

	warrantClient := warrant.New(warrant.Config{
		Host:          uaaBaseURL,
		SkipVerifySSL: skipSSLValidation,
	})
	uaaClient := &uaa_client.Client{
		Service: warrantClient.Clients,
		User:    uaaClientName,
		Secret:  uaaSecret,
	}
	rainmakerClient := rainmaker.NewClient(rainmaker.Config{
		Host: ccClientHost,
	})
	ccClient := cc_client.Client{
		Org:   rainmakerClient.Organizations,
		Space: rainmakerClient.Spaces,
		App:   rainmakerClient.Applications,
		UAA:   uaaClient,
	}
	ducatiDaemonClient := client.New(ducatiAPI, http.DefaultClient)
	httpResolver := &resolver.HTTPResolver{
		Logger:       logger,
		Suffix:       ducatiSuffix,
		DaemonClient: ducatiDaemonClient,
		CCClient:     &ccClient,
	}
	resolverMuxer := dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		if strings.HasSuffix(request.Question[0].Name, fmt.Sprintf("%s.", ducatiSuffix)) {
			httpResolver.ServeDNS(w, request)
		} else {
			forwardingResolver.ServeDNS(w, request)
		}
	})

	udpAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	if err != nil {
		log.Fatalf("invalid listen address %s: %s", listenAddress, err)
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
