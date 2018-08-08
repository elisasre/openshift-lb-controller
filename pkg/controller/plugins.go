/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	"log"
	"os"
	"strings"
	"sync"

	v1 "github.com/openshift/api/route/v1"
)

// ProviderInterface is an abstract, pluggable interface for different loadbalancers.
type ProviderInterface interface {
	// Initialize initilizes new provider
	Initialize()
	// creates new loadbalancer pool
	CreatePool(name string, port string) error
	// adds new member to pool
	AddPoolMember(membername string, name string, port string) error
	// modifies loadbalancer pool
	ModifyPool(name string, port string, loadBalancingMethod string, pga int, maintenance bool, prio int) error
	// creates new monitor
	CreateMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error
	// modifies monitor
	ModifyMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error
	// adds monitor to pool
	AddMonitorToPool(name string, port string) error
	// delete pool member
	DeletePoolMember(membername string, name string, port string) error
	// checks pool members and if 0 members left in pool, delete monitor and delete pool
	CheckAndClean(name string, port string)
	// executed before something is updated. Can be used for instance to checking active member of the HA lb
	PreUpdate()
	// executed after something is updated. Can be used for instance to configuration sync
	PostUpdate()
	// returns hosts which should be removed
	CheckPools(routes []v1.Route, hosttowatch string, membername string) map[string]bool
	// testing purposes
	Calls() []string
	CleanCalls()
}

var (
	providersMutex sync.Mutex
	providers      = make(map[string]ProviderInterface)
)

// RegisterProvider registers new load balancer provider
func RegisterProvider(name string, cloud ProviderInterface) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	log.Printf("Registered provider %q", name)
	providers[name] = cloud
}

func getProvider(name string) ProviderInterface {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	f, found := providers[name]
	if !found {
		return nil
	}
	return f
}

// InitProvider returns load balancer providerinterface
func (c *RouteController) InitProvider() ProviderInterface {
	name := strings.ToLower(os.Getenv("PROVIDER"))
	cloud := getProvider(name)
	log.Printf("Using provider %s", name)
	return cloud
}
