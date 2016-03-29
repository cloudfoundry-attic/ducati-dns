package main

import (
	"flag"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
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

	members := grouper.Members{
		{"dns_server", forwardingResolver},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	err := <-monitor.Wait()
	if err != nil {
		log.Fatalf("daemon terminated: %s", err)
	}
}
