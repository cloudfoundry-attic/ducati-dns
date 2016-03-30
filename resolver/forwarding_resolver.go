package resolver

import (
	"fmt"
	"sync"
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
	r := make(chan *dns.Msg)
	var wg sync.WaitGroup
	serverLookup := func(nameserver string) {
		resp, _, err := h.Exchanger.Exchange(request, nameserver)
		if err != nil {
			fmt.Errorf("Exchange err: %s", err)
			wg.Done()
			return
		}
		wg.Done()
		r <- resp
	}

	for _, server := range h.Servers {
		go serverLookup(server)
		wg.Add(1)
	}

	wg.Wait()
	select {
	case resp := <-r:
		w.WriteMsg(resp)
		return
	default:
		m := &dns.Msg{}
		m.SetReply(request)
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}
}
