package main

import (
	"flag"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/miekg/dns"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func main() {
	var server string
	flag.StringVar(&server, "server", "", "")
	flag.Parse()

	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Servers:   []string{server},
	}

	dnsRunner := &runner.Runner{
		DNSServer: &dns.Server{
			Addr:    "127.0.0.1:9999",
			Net:     "udp",
			Handler: forwardingResolver,
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
