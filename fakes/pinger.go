// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"database/sql/driver"
	"sync"
)

type FakePinger struct {
	PingStub        func(ctx context.Context) error
	pingMutex       sync.RWMutex
	pingArgsForCall []struct {
		ctx context.Context
	}
	pingReturns struct {
		result1 error
	}
	pingReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePinger) Ping(ctx context.Context) error {
	fake.pingMutex.Lock()
	ret, specificReturn := fake.pingReturnsOnCall[len(fake.pingArgsForCall)]
	fake.pingArgsForCall = append(fake.pingArgsForCall, struct {
		ctx context.Context
	}{ctx})
	fake.recordInvocation("Ping", []interface{}{ctx})
	fake.pingMutex.Unlock()
	if fake.PingStub != nil {
		return fake.PingStub(ctx)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.pingReturns.result1
}

func (fake *FakePinger) PingCallCount() int {
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	return len(fake.pingArgsForCall)
}

func (fake *FakePinger) PingArgsForCall(i int) context.Context {
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	return fake.pingArgsForCall[i].ctx
}

func (fake *FakePinger) PingReturns(result1 error) {
	fake.PingStub = nil
	fake.pingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePinger) PingReturnsOnCall(i int, result1 error) {
	fake.PingStub = nil
	if fake.pingReturnsOnCall == nil {
		fake.pingReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pingReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakePinger) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakePinger) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ driver.Pinger = new(FakePinger)