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
	Logger    lager.Logger
	Exchanger exchanger
	Server    string
}

func (h *ForwardingResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	logger := h.Logger.Session("serve-dns", lager.Data{"name": request.Question[0].Name})
	logger.Info("resolving")
	defer logger.Info("resolve-complete")

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
		logger.Info("nil-response")
		m := &dns.Msg{}
		m.SetReply(request)
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	logger.Info("response", lager.Data{"answer": resp.Answer})

	w.WriteMsg(resp)
}
