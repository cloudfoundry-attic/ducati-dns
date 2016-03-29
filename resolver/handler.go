package resolver

import (
	"os"
	"time"

	"github.com/miekg/dns"
)

//go:generate counterfeiter -o ../fakes/handler.go --fake-name Handler . handler
type handler interface {
	dns.Handler
}

//go:generate counterfeiter -o ../fakes/response_writer.go --fake-name ResponseWriter . responseWriter
type responseWriter interface {
	dns.ResponseWriter
}

//go:generate counterfeiter -o ../fakes/exchanger.go --fake-name Exchanger . exchanger
type exchanger interface {
	Exchange(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error)
}

type ForwardingResolver struct {
	Exchanger exchanger
	Servers   []string
}

func (h *ForwardingResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	for _, server := range h.Servers {
		resp, _, err := h.Exchanger.Exchange(request, server)
		if err != nil {
			panic(err)
		}
		w.WriteMsg(resp)
		return
	}

	m := &dns.Msg{}
	m.SetReply(request)
	m.SetRcode(request, dns.RcodeNameError)
	w.WriteMsg(m)
}

func (h *ForwardingResolver) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
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
