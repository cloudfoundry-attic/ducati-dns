// This file was generated by counterfeiter
package fakes

import (
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Exchanger struct {
	ExchangeStub        func(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error)
	exchangeMutex       sync.RWMutex
	exchangeArgsForCall []struct {
		m *dns.Msg
		a string
	}
	exchangeReturns struct {
		result1 *dns.Msg
		result2 time.Duration
		result3 error
	}
}

func (fake *Exchanger) Exchange(m *dns.Msg, a string) (r *dns.Msg, rtt time.Duration, err error) {
	fake.exchangeMutex.Lock()
	fake.exchangeArgsForCall = append(fake.exchangeArgsForCall, struct {
		m *dns.Msg
		a string
	}{m, a})
	fake.exchangeMutex.Unlock()
	if fake.ExchangeStub != nil {
		return fake.ExchangeStub(m, a)
	} else {
		return fake.exchangeReturns.result1, fake.exchangeReturns.result2, fake.exchangeReturns.result3
	}
}

func (fake *Exchanger) ExchangeCallCount() int {
	fake.exchangeMutex.RLock()
	defer fake.exchangeMutex.RUnlock()
	return len(fake.exchangeArgsForCall)
}

func (fake *Exchanger) ExchangeArgsForCall(i int) (*dns.Msg, string) {
	fake.exchangeMutex.RLock()
	defer fake.exchangeMutex.RUnlock()
	return fake.exchangeArgsForCall[i].m, fake.exchangeArgsForCall[i].a
}

func (fake *Exchanger) ExchangeReturns(result1 *dns.Msg, result2 time.Duration, result3 error) {
	fake.ExchangeStub = nil
	fake.exchangeReturns = struct {
		result1 *dns.Msg
		result2 time.Duration
		result3 error
	}{result1, result2, result3}
}