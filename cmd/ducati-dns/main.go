package main

import (
	"log"
	"net"
	"os"

	"github.com/miekg/dns"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

type handler struct{}

func (h *handler) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	q := request.Question[0]

	m := &dns.Msg{}
	m.SetReply(request)
	rr_header := dns.RR_Header{
		Name:   q.Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    1,
	}

	a := &dns.A{rr_header, net.ParseIP("93.184.216.34")}

	m.Answer = []dns.RR{a}
	w.WriteMsg(m)
}

func (h *handler) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	udpServer := &dns.Server{
		Addr:              "127.0.0.1:9999",
		Net:               "udp",
		Handler:           h,
		NotifyStartedFunc: func() { close(ready) },
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- udpServer.ListenAndServe()
	}()

	for {
		select {
		case err := <-errCh:
			return err

		case <-signals:
			return udpServer.Shutdown()
		}
	}
}

func main() {
	members := grouper.Members{
		{"dns_server", &handler{}},
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	err := <-monitor.Wait()
	if err != nil {
		log.Fatalf("daemon terminated: %s", err)
	}
}
