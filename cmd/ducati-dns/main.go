package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func main() {
	var (
		server       string
		ducatiSuffix string
		ducatiAPI    string
	)

	flag.StringVar(&server, "server", "", "Single DNS server to forward queries to")
	flag.StringVar(&ducatiSuffix, "ducatiSuffix", "", "suffix for lookups on the overlay network")
	flag.StringVar(&ducatiAPI, "ducatiAPI", "", "URL for the ducati API")

	var listenAddress string
	flag.StringVar(&listenAddress, "listenAddress", "127.0.0.1:53", "Host and port to listen for queries on")
	flag.Parse()

	if server == "" {
		fmt.Fprintf(os.Stderr, "missing required arg: server")
		os.Exit(1)
	}
	if ducatiSuffix == "" {
		fmt.Fprintf(os.Stderr, "missing required arg: ducatiSuffix")
		os.Exit(1)
	}
	if ducatiAPI == "" {
		fmt.Fprintf(os.Stderr, "missing required arg: ducatiAPI")
		os.Exit(1)
	}

	logger := lager.NewLogger("ducati-dns")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Server:    server,
		Logger:    logger,
	}

	ducatiDaemonClient := client.New(ducatiAPI, http.DefaultClient)
	httpResolver := &resolver.HTTPResolver{
		Logger:       logger,
		Suffix:       ducatiSuffix,
		DaemonClient: ducatiDaemonClient,
	}
	resolverMuxer := dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		if strings.HasSuffix(request.Question[0].Name, fmt.Sprintf("%s.", ducatiSuffix)) {
			httpResolver.ServeDNS(w, request)
		} else {
			forwardingResolver.ServeDNS(w, request)
		}
	})

	dnsRunner := &runner.Runner{
		DNSServer: &dns.Server{
			Addr:    listenAddress,
			Net:     "udp",
			Handler: resolverMuxer,
		},
	}

	members := grouper.Members{
		{"dns_runner", dnsRunner},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	err := <-monitor.Wait()
	if err != nil {
		log.Fatalf("daemon terminated: %s", err)
	}
}
