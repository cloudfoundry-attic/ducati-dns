// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/ducati-daemon/models"
)

type DucatiDaemonClient struct {
	GetContainerStub        func(containerID string) (models.Container, error)
	getContainerMutex       sync.RWMutex
	getContainerArgsForCall []struct {
		containerID string
	}
	getContainerReturns struct {
		result1 models.Container
		result2 error
	}
}

func (fake *DucatiDaemonClient) GetContainer(containerID string) (models.Container, error) {
	fake.getContainerMutex.Lock()
	fake.getContainerArgsForCall = append(fake.getContainerArgsForCall, struct {
		containerID string
	}{containerID})
	fake.getContainerMutex.Unlock()
	if fake.GetContainerStub != nil {
		return fake.GetContainerStub(containerID)
	} else {
		return fake.getContainerReturns.result1, fake.getContainerReturns.result2
	}
}

func (fake *DucatiDaemonClient) GetContainerCallCount() int {
	fake.getContainerMutex.RLock()
	defer fake.getContainerMutex.RUnlock()
	return len(fake.getContainerArgsForCall)
}

func (fake *DucatiDaemonClient) GetContainerArgsForCall(i int) string {
	fake.getContainerMutex.RLock()
	defer fake.getContainerMutex.RUnlock()
	return fake.getContainerArgsForCall[i].containerID
}

func (fake *DucatiDaemonClient) GetContainerReturns(result1 models.Container, result2 error) {
	fake.GetContainerStub = nil
	fake.getContainerReturns = struct {
		result1 models.Container
		result2 error
	}{result1, result2}
}
