package resolver

import (
	"runtime"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/lib/debug"
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
	runtime.LockOSThread()

	suffix := dns.Fqdn(m.Suffix)
	name := request.Question[0].Name

	logger := m.Logger.Session("serve-dns", lager.Data{"name": name})

	logger.Info("resolving", debug.NetNS())
	defer logger.Info("complete", debug.NetNS())

	if m.Suffix != "" && strings.HasSuffix(name, suffix) {
		m.SuffixPresentHandler.ServeDNS(w, request)
	} else {
		m.DefaultHandler.ServeDNS(w, request)
	}
}
