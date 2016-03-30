package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/miekg/dns"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

type servers []string

func (s *servers) String() string {
	return fmt.Sprint(*s)
}

func (s *servers) Set(value string) error {
	for _, server := range strings.Split(value, ",") {
		*s = append(*s, server)
	}
	return nil
}

func main() {
	var serverList servers
	flag.Var(&serverList, "server", "")
	flag.Parse()

	forwardingResolver := &resolver.ForwardingResolver{
		Exchanger: &dns.Client{Net: "udp"},
		Servers:   serverList,
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
