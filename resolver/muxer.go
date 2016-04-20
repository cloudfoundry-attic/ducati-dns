package resolver

import (
	"strings"

	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

type Muxer struct {
	Logger               lager.Logger
	Suffix               string
	SuffixPresentHandler dns.Handler
	DefaultHandler       dns.Handler
}

func (m *Muxer) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	suffix := dns.Fqdn(m.Suffix)
	name := request.Question[0].Name

	m.Logger.Info("serve-dns", lager.Data{"name": name})

	if m.Suffix != "" && strings.HasSuffix(name, suffix) {
		m.SuffixPresentHandler.ServeDNS(w, request)
	} else {
		m.DefaultHandler.ServeDNS(w, request)
	}
}
