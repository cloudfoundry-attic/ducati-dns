package resolver

import (
	"time"

	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
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
	Server    string
	Logger    lager.Logger
}

func (h *ForwardingResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	resp, _, err := h.Exchanger.Exchange(request, h.Server)
	if err != nil {
		h.Logger.Error("Serve DNS Exchange", err)

		m := &dns.Msg{}
		m.SetReply(request)
		m.SetRcode(request, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	if resp == nil {
		m := &dns.Msg{}
		m.SetReply(request)
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	w.WriteMsg(resp)
}
