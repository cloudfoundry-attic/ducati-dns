// This file was generated by counterfeiter
package fakes

import "sync"

type CCClient struct {
	GetAppGuidStub        func(appName string, space string, org string) (string, error)
	getAppGuidMutex       sync.RWMutex
	getAppGuidArgsForCall []struct {
		appName string
		space   string
		org     string
	}
	getAppGuidReturns struct {
		result1 string
		result2 error
	}
}

func (fake *CCClient) GetAppGuid(appName string, space string, org string) (string, error) {
	fake.getAppGuidMutex.Lock()
	fake.getAppGuidArgsForCall = append(fake.getAppGuidArgsForCall, struct {
		appName string
		space   string
		org     string
	}{appName, space, org})
	fake.getAppGuidMutex.Unlock()
	if fake.GetAppGuidStub != nil {
		return fake.GetAppGuidStub(appName, space, org)
	} else {
		return fake.getAppGuidReturns.result1, fake.getAppGuidReturns.result2
	}
}

func (fake *CCClient) GetAppGuidCallCount() int {
	fake.getAppGuidMutex.RLock()
	defer fake.getAppGuidMutex.RUnlock()
	return len(fake.getAppGuidArgsForCall)
}

func (fake *CCClient) GetAppGuidArgsForCall(i int) (string, string, string) {
	fake.getAppGuidMutex.RLock()
	defer fake.getAppGuidMutex.RUnlock()
	return fake.getAppGuidArgsForCall[i].appName, fake.getAppGuidArgsForCall[i].space, fake.getAppGuidArgsForCall[i].org
}

func (fake *CCClient) GetAppGuidReturns(result1 string, result2 error) {
	fake.GetAppGuidStub = nil
	fake.getAppGuidReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}
