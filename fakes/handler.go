// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/miekg/dns"
)

type Handler struct {
	ServeDNSStub        func(w dns.ResponseWriter, r *dns.Msg)
	serveDNSMutex       sync.RWMutex
	serveDNSArgsForCall []struct {
		w dns.ResponseWriter
		r *dns.Msg
	}
}

func (fake *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	fake.serveDNSMutex.Lock()
	fake.serveDNSArgsForCall = append(fake.serveDNSArgsForCall, struct {
		w dns.ResponseWriter
		r *dns.Msg
	}{w, r})
	fake.serveDNSMutex.Unlock()
	if fake.ServeDNSStub != nil {
		fake.ServeDNSStub(w, r)
	}
}

func (fake *Handler) ServeDNSCallCount() int {
	fake.serveDNSMutex.RLock()
	defer fake.serveDNSMutex.RUnlock()
	return len(fake.serveDNSArgsForCall)
}

func (fake *Handler) ServeDNSArgsForCall(i int) (dns.ResponseWriter, *dns.Msg) {
	fake.serveDNSMutex.RLock()
	defer fake.serveDNSMutex.RUnlock()
	return fake.serveDNSArgsForCall[i].w, fake.serveDNSArgsForCall[i].r
}
