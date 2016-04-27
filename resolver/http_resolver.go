package resolver

import (
	"net"
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	"github.com/miekg/dns"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../fakes/ducati_daemon_client.go --fake-name DucatiDaemonClient . ducatiDaemonClient
type ducatiDaemonClient interface {
	ListContainers() ([]models.Container, error)
}

type Config struct {
	DucatiSuffix string
	DucatiAPI    string
}

func NewHTTPResolver(logger lager.Logger, config Config) *HTTPResolver {
	ducatiDaemonClient := client.New(
		config.DucatiAPI,
		http.DefaultClient,
	)
	return &HTTPResolver{
		Logger:       logger,
		Suffix:       config.DucatiSuffix,
		DaemonClient: ducatiDaemonClient,
	}
}

type HTTPResolver struct {
	DaemonClient ducatiDaemonClient
	TTL          int
	Suffix       string
	Logger       lager.Logger
}

func (r *HTTPResolver) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	m := &dns.Msg{}

	requestedName := request.Question[0].Name
	fullyQualifiedSuffix := "." + r.Suffix + "."
	appGuid := strings.TrimSuffix(requestedName, fullyQualifiedSuffix)
	if appGuid == requestedName {
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		r.Logger.Info("unknown-name", lager.Data{"requested_name": requestedName})
		return
	}

	containers, err := r.DaemonClient.ListContainers()
	if err != nil {
		m.SetRcode(request, dns.RcodeServerFailure)
		w.WriteMsg(m)
		r.Logger.Error("ducati-client-error", err)
		return
	}

	var container models.Container
	for _, c := range containers {
		if c.App == appGuid {
			container = c
		}
	}

	if container == (models.Container{}) {
		m.SetRcode(request, dns.RcodeNameError)
		w.WriteMsg(m)
		r.Logger.Info("record-not-found", lager.Data{"requested_name": requestedName})
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
