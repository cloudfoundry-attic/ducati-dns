// This file was generated by counterfeiter
package fakes

import "sync"

type DNSServer struct {
	ListenAndServeStub        func() error
	listenAndServeMutex       sync.RWMutex
	listenAndServeArgsForCall []struct{}
	listenAndServeReturns     struct {
		result1 error
	}
	ShutdownStub        func() error
	shutdownMutex       sync.RWMutex
	shutdownArgsForCall []struct{}
	shutdownReturns     struct {
		result1 error
	}
}

func (fake *DNSServer) ListenAndServe() error {
	fake.listenAndServeMutex.Lock()
	fake.listenAndServeArgsForCall = append(fake.listenAndServeArgsForCall, struct{}{})
	fake.listenAndServeMutex.Unlock()
	if fake.ListenAndServeStub != nil {
		return fake.ListenAndServeStub()
	} else {
		return fake.listenAndServeReturns.result1
	}
}

func (fake *DNSServer) ListenAndServeCallCount() int {
	fake.listenAndServeMutex.RLock()
	defer fake.listenAndServeMutex.RUnlock()
	return len(fake.listenAndServeArgsForCall)
}

func (fake *DNSServer) ListenAndServeReturns(result1 error) {
	fake.ListenAndServeStub = nil
	fake.listenAndServeReturns = struct {
		result1 error
	}{result1}
}

func (fake *DNSServer) Shutdown() error {
	fake.shutdownMutex.Lock()
	fake.shutdownArgsForCall = append(fake.shutdownArgsForCall, struct{}{})
	fake.shutdownMutex.Unlock()
	if fake.ShutdownStub != nil {
		return fake.ShutdownStub()
	} else {
		return fake.shutdownReturns.result1
	}
}

func (fake *DNSServer) ShutdownCallCount() int {
	fake.shutdownMutex.RLock()
	defer fake.shutdownMutex.RUnlock()
	return len(fake.shutdownArgsForCall)
}

func (fake *DNSServer) ShutdownReturns(result1 error) {
	fake.ShutdownStub = nil
	fake.shutdownReturns = struct {
		result1 error
	}{result1}
}