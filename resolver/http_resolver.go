package resolver

import (
	"net"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	"github.com/cloudfoundry-incubator/ducati-dns/cc_client"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/ducati_daemon_client.go --fake-name DucatiDaemonClient . ducatiDaemonClient
type ducatiDaemonClient interface {
	GetContainer(containerID string) (models.Container, error)
}

//go:generate counterfeiter -o ../fakes/cc_client.go --fake-name CCClient . ccClient
type ccClient interface {
	GetAppGuid(appName string, space string, org string) (string, error)
}

type HTTPResolver struct {
	DaemonClient ducatiDaemonClient
	CCClient     ccClient
	TTL          int
	Suffix       string
	Logger       lager.Logger
}

func (r *HTTPResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	m := &dns.Msg{}

	requestedName := request.Question[0].Name
	fullyQualifiedSuffix := "." + r.Suffix + "."
	appSpaceOrg := strings.TrimSuffix(requestedName, fullyQualifiedSuffix)
	if appSpaceOrg == requestedName {
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		r.Logger.Info("unknown-name", lager.Data{"requested_name": requestedName})
		return
	}

	domainParts := strings.Split(appSpaceOrg, ".")
	if len(domainParts) != 3 {
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		r.Logger.Info("invalid-domain", lager.Data{"requested_name": requestedName})
		return
	}

	containerName, err := r.CCClient.GetAppGuid(domainParts[0], domainParts[1], domainParts[2])
	if err != nil {
		if err == cc_client.DomainNotFoundError {
			m.SetRcode(request, dns.RcodeNameError)
			w.WriteMsg(m)
			r.Logger.Info("domain-not-found", lager.Data{"requested_name": requestedName})
			return
		}
		m.SetRcode(request, dns.RcodeServerFailure)
		w.WriteMsg(m)
		r.Logger.Error("cloud-controller-client-error", err)
		return
	}

	container, err := r.DaemonClient.GetContainer(containerName)
	if err != nil {
		if err == client.RecordNotFoundError {
			m.SetRcode(request, dns.RcodeNameError)
			w.WriteMsg(m)
			r.Logger.Info("record-not-found", lager.Data{"requested_name": requestedName})
			return
		}
		m.SetRcode(request, dns.RcodeServerFailure)
		w.WriteMsg(m)
		r.Logger.Error("ducati-client-error", err)
		return
	}

	m.SetReply(request)

	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   requestedName,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(r.TTL),
			},
			A: net.ParseIP(container.IP)},
	}
	w.WriteMsg(m)
}
