package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"os"

	"github.com/cloudfoundry-incubator/ducati-dns/resolver"
	"github.com/cloudfoundry-incubator/ducati-dns/runner"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func validate(c resolver.Config) error {
	if c.DucatiSuffix == "" {
		return errors.New("missing required arg: ducatiSuffix")
	}
	if c.DucatiAPI == "" {
		return errors.New("missing required arg: ducatiAPI")
	}

	return nil
}

func main() {
	var (
		config            resolver.Config
		externalDNSServer string
		listenAddress     string
	)

	flag.StringVar(&externalDNSServer, "server", "", "Single DNS server to forward queries to")
	flag.StringVar(&config.DucatiSuffix, "ducatiSuffix", "", "suffix for lookups on the overlay network")
	flag.StringVar(&config.DucatiAPI, "ducatiAPI", "", "URL for the ducati API")
	flag.StringVar(&listenAddress, "listenAddress", "127.0.0.1:53", "Host and port to listen for queries on")
	flag.Parse()

	if err := validate(config); err != nil {
		log.Fatalf("validate: %s", err)
	}

	if externalDNSServer == "" {
		log.Fatalf("missing required arg: server")
	}

	logger := lager.NewLogger("ducati-dns")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))

	udpAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	if err != nil {
		log.Fatalf("invalid listen address %s: %s", listenAddress, err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("listen: %s", err)
	}
	defer udpConn.Close()

	dnsRunner := runner.New(logger, config, externalDNSServer, udpConn, nil)

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
