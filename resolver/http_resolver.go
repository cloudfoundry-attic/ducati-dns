package resolver

import (
	"net"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/ducati_daemon_client.go --fake-name DucatiDaemonClient . ducatiDaemonClient
type ducatiDaemonClient interface {
	GetContainer(containerID string) (models.Container, error)
}

type HTTPResolver struct {
	DaemonClient ducatiDaemonClient
	TTL          int
	Suffix       string
	Logger       lager.Logger
}

func (r *HTTPResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	requestedName := request.Question[0].Name
	fullyQualifiedSuffix := "." + r.Suffix + "."
	containerName := strings.TrimSuffix(requestedName, fullyQualifiedSuffix)
	if containerName == requestedName {
		m := &dns.Msg{}
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		r.Logger.Info("unknown-name", lager.Data{"requested_name": requestedName})
		return
	}

	container, err := r.DaemonClient.GetContainer(containerName)
	if err != nil {
		m := &dns.Msg{}
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

	m := &dns.Msg{}
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
