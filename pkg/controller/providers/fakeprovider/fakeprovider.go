/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package fakeprovider

import (
	v1 "github.com/openshift/api/route/v1"
	"sync"
)

// Fakeprovider is an implementation of Interface for fakeprovider which helps testing.
type Fakeprovider struct {
	calls       []string
	addCallLock sync.Mutex
}

func init() {
}

// NewFakeProvider returns new fakeprovider for testing purposes
func NewFakeProvider() *Fakeprovider {
	fake := Fakeprovider{
		calls: []string{},
	}
	return &fake
}

func (f *Fakeprovider) addCall(desc string) {
	f.addCallLock.Lock()
	defer f.addCallLock.Unlock()
	f.calls = append(f.calls, desc)
}

// Initialize initilizes new provider
func (f *Fakeprovider) Initialize() {
	f.addCall("Initialize")
}

// AddPoolMember adds new member to pool
func (f *Fakeprovider) AddPoolMember(membername string, name string, port string) error {
	f.addCall("AddPoolMember")
	return nil
}

// CreatePool creates new loadbalancer pool
func (f *Fakeprovider) CreatePool(name string, port string) error {
	f.addCall("CreatePool")
	return nil
}

// ModifyPool modifies loadbalancer pool
func (f *Fakeprovider) ModifyPool(name string, port string, loadBalancingMethod string, pga int) error {
	f.addCall("ModifyPool")
	return nil
}

// CreateMonitor creates new monitor
func (f *Fakeprovider) CreateMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error {
	f.addCall("CreateMonitor")
	return nil
}

// ModifyMonitor modifies monitor
func (f *Fakeprovider) ModifyMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error {
	f.addCall("ModifyMonitor")
	return nil
}

// AddMonitorToPool adds monitor to pool
func (f *Fakeprovider) AddMonitorToPool(name string, port string) error {
	f.addCall("AddMonitorToPool")
	return nil
}

// DeletePoolMember delete pool member
func (f *Fakeprovider) DeletePoolMember(membername string, name string, port string) error {
	f.addCall("DeletePoolMember")
	return nil
}

// CheckAndClean checks pool members and if 0 members left in pool, delete monitor and delete pool
func (f *Fakeprovider) CheckAndClean(name string, port string) {
	f.addCall("CheckAndClean")
}

// CheckPools compares current load balancer setup and what routes we have. It returns list of pools which should be removed
func (f *Fakeprovider) CheckPools(routes []v1.Route, hosttowatch string, membername string) []string {
	f.addCall("CheckPools")
	return nil
}

// PreUpdate is executed before updating anything
func (f *Fakeprovider) PreUpdate() {
	f.addCall("PreUpdate")
}

// PostUpdate is executed after updating
func (f *Fakeprovider) PostUpdate() {
	f.addCall("PostUpdate")
}

// Calls returns list of methodcalls
func (f *Fakeprovider) Calls() []string {
	return f.calls
}

// CleanCalls cleans calls
func (f *Fakeprovider) CleanCalls() {
	f.calls = []string{}
}
