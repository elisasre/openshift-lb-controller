/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package f5

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ElisaOyj/openshift-lb-controller/pkg/common"
	"github.com/ElisaOyj/openshift-lb-controller/pkg/controller"
	"github.com/getsentry/raven-go"
	v1 "github.com/openshift/api/route/v1"
	bigip "github.com/scottdware/go-bigip"
)

const providerName = "f5"

// ProviderF5 is an implementation of Interface for F5.
type ProviderF5 struct {
	session      *bigip.BigIP
	Clusteralias string
	username     string
	password     string
	addresses    []string
	currentaddr  int
	groupname    string
	partition    string
}

func init() {
	controller.RegisterProvider(providerName, NewProviderF5())
}

// NewProviderF5 returns new f5 provider for testing purposes
func NewProviderF5() *ProviderF5 {
	f5 := ProviderF5{
		currentaddr: 0,
		groupname:   "cluster",
		partition:   "Common",
	}
	return &f5
}

func alreadyExist(err error, partition string) bool {
	compareTxt := fmt.Sprintf("already exists in partition %s.", partition)
	if strings.HasSuffix(err.Error(), compareTxt) {
		return true
	}
	return false
}

func getNameWithPool(partition string, name string) string {
	return fmt.Sprintf("/%s/%s", partition, name)
}

// Initialize initilizes new provider
func (f5 *ProviderF5) Initialize() {
	address := os.Getenv("F5_ADDR")
	if len(address) == 0 {
		err := errors.New("F5_ADDR environment variable needed")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}

	groupname := os.Getenv("F5_CLUSTERGROUP")
	if len(groupname) != 0 {
		f5.groupname = groupname
	}
	f5.addresses = strings.Split(address, ",")

	username := os.Getenv("F5_USER")
	if len(username) == 0 {
		err := errors.New("F5_USER environment variable needed")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}
	password := os.Getenv("F5_PASSWORD")
	if len(password) == 0 {
		err := errors.New("F5_PASSWORD environment variable needed")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}
	f5.username = username
	f5.password = password

	partition := os.Getenv("PARTITION")
	if len(partition) > 0 {
		f5.partition = partition
	}

	f5.session = bigip.NewSession(f5.addresses[0], f5.username, f5.password, nil)
}

// CreatePool creates new loadbalancer pool
func (f5 *ProviderF5) CreatePool(name string, port string) error {
	f5Pool := &bigip.Pool{
		Name:              name + "_" + port,
		Partition:         f5.partition,
		ServiceDownAction: "reset",
	}
	err := f5.session.AddPool(f5Pool)
	if err != nil {
		if !alreadyExist(err, f5.partition) {
			return err
		}
	}
	return nil
}

// AddPoolMember adds new member to pool
func (f5 *ProviderF5) AddPoolMember(membername string, name string, port string) error {
	f5.Clusteralias = membername
	err := f5.session.AddPoolMember(getNameWithPool(f5.partition, name+"_"+port), getNameWithPool(f5.partition, membername+":"+port))
	if err != nil {
		if !alreadyExist(err, f5.partition) {
			return err
		}
	}
	return nil
}

func (f5 *ProviderF5) modifyMember(name string, port string, maintenance bool, prio int) {
	// we need use this because getpoolmember is not working correctly
	members, err := f5.session.PoolMembers(getNameWithPool(f5.partition, name+"_"+port))
	if err != nil {
		log.Printf("error in getpoolmembers %v", err)
		return
	}
	for _, item := range members.PoolMembers {
		if item.Name == f5.Clusteralias+":"+port {
			config := &bigip.PoolMember{
				FullPath:      item.FullPath,
				PriorityGroup: prio,
				Partition:     f5.partition,
			}
			if maintenance {
				config.Session = "user-disabled"
				log.Printf("setting poolmember %s in pool %s_%s to disabled", f5.Clusteralias, name, port)
			} else {
				config.Session = "user-enabled"
				log.Printf("setting poolmember %s in pool %s_%s to enabled", f5.Clusteralias, name, port)
			}
			err = f5.session.PatchPoolMember(getNameWithPool(f5.partition, name+"_"+port), config)
			if err != nil {
				log.Printf("error in modifyMember %v", err)
			}
			break
		}
	}
}

// ModifyPool modifies loadbalancer pool
func (f5 *ProviderF5) ModifyPool(name string, port string, loadBalancingMethod string, pga int, maintenance bool, prio int, role string) error {
	pool, err := f5.session.GetPool(getNameWithPool(f5.partition, name+"_"+port))
	if err != nil {
		return err
	}
	targetmode := loadBalancingMethod
	if len(loadBalancingMethod) == 0 {
		targetmode = "round-robin"
	}
	log.Printf("changing pool %s loadbalancingmode to %s", name+"_"+port, targetmode)
	pool.LoadBalancingMode = targetmode
	if strings.ToLower(role) == "active" || strings.ToLower(role) == "standby" {
		log.Printf("modifying slow ramp time pool to 0 %s", name+"_"+port)
		pool.SlowRampTime = f5.session.IntToPointer(0)
		pga = 1
		if strings.ToLower(role) == "active" {
			prio = 20
		} else if strings.ToLower(role) == "standby" {
			prio = 15
		}
	} else {
		log.Printf("modifying slow ramp time pool to 10 %s", name+"_"+port)
		pool.SlowRampTime = f5.session.IntToPointer(10)
	}
	log.Printf("changing pool %s pga to %d", name+"_"+port, pga)
	pool.MinActiveMembers = pga
	f5.modifyMember(name, port, maintenance, prio)
	// override servicedownaction to reset
	log.Printf("changing pool serviceaction down to reset %s", name+"_"+port)
	pool.ServiceDownAction = "reset"
	err = f5.session.ModifyPool(name+"_"+port, pool)
	if err != nil {
		return err
	}
	return nil
}

// CreateMonitor creates new monitor
func (f5 *ProviderF5) CreateMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	err := f5.session.CreateMonitor(getNameWithPool(f5.partition, host+"_"+port), scheme, interval, timeout, httpMethod+" "+uri+" HTTP/1.1\r\nHost:"+host+"  \r\nConnection: Close\r\n\r\n", "^HTTP.1.(0|1) ([2|3]0[0-9])", scheme)
	if err != nil {
		if !alreadyExist(err, f5.partition) {
			return err
		}
	}
	return nil
}

// ModifyMonitor modifies monitor
func (f5 *ProviderF5) ModifyMonitor(host string, port string, uri string, httpMethod string, interval int, timeout int) error {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	config := &bigip.Monitor{
		Interval:   interval,
		Timeout:    timeout,
		SendString: httpMethod + " " + uri + " HTTP/1.1\r\nHost:" + host + "  \r\nConnection: Close\r\n\r\n",
		Partition:  f5.partition,
	}
	err := f5.session.PatchMonitor(getNameWithPool(f5.partition, host+"_"+port), scheme, config)
	if err != nil {
		return err
	}
	return nil
}

// AddMonitorToPool adds monitor to pool
func (f5 *ProviderF5) AddMonitorToPool(name string, port string) error {
	err := f5.session.AddMonitorToPool(name+"_"+port, getNameWithPool(f5.partition, name+"_"+port))
	if err != nil {
		if !alreadyExist(err, f5.partition) {
			return err
		}
	}
	return nil
}

// DeletePoolMember delete pool member
func (f5 *ProviderF5) DeletePoolMember(membername string, poolname string, poolport string) error {
	return f5.session.DeletePoolMember(getNameWithPool(f5.partition, poolname+"_"+poolport), membername+":"+poolport)
}

// CheckAndClean checks pool members and if 0 members left in pool, delete monitor and delete pool
func (f5 *ProviderF5) CheckAndClean(name string, port string) {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	members, err := f5.session.PoolMembers(getNameWithPool(f5.partition, name+"_"+port))
	if err != nil {
		log.Printf("error retrieving poolmembers %s %v", name+"_"+port, err)
	}
	if len(members.PoolMembers) == 0 {
		f5name := getNameWithPool(f5.partition, name+"_"+port)
		err = f5.session.DeletePool(f5name)
		if err != nil {
			log.Printf("error delete pool %s %v", f5name, err)
		}
		err = f5.session.DeleteMonitor(f5name, scheme)
		if err != nil {
			log.Printf("error delete monitor %s %v", f5name, err)
		}
	}
}

func (f5 *ProviderF5) poolMemberExist(pool bigip.Pool, membername string) bool {
	members, err := f5.session.PoolMembers(getNameWithPool(f5.partition, pool.Name))
	if err != nil {
		log.Printf("error in poolmembers %v", err)
		return false
	}

	for _, member := range members.PoolMembers {
		s := strings.Split(member.Name, ":")
		clustername := s[0]
		if clustername == membername {
			return true
		}
	}
	return false
}

func (f5 *ProviderF5) getPools() (*bigip.Pools, error) {
	var filteredPools *bigip.Pools
	filteredPools = &bigip.Pools{}
	pools, err := f5.session.Pools()
	if err != nil {
		return filteredPools, err
	}
	for _, pool := range pools.Pools {
		if len(f5.partition) == 0 && pool.Partition == "Common" {
			filteredPools.Pools = append(filteredPools.Pools, pool)
		} else if pool.Partition == f5.partition {
			filteredPools.Pools = append(filteredPools.Pools, pool)
		}
	}
	return filteredPools, err
}

// CheckPools compares current load balancer setup and what routes we have. It returns list of pools which should be removed
func (f5 *ProviderF5) CheckPools(routes []v1.Route, hosttowatch string, membername string) map[string]bool {
	hosts := map[string]bool{}
	pools, err := f5.getPools()
	if err != nil {
		log.Printf("error fetching pool %v", err)
		return hosts
	}
	for _, pool := range pools.Pools {
		if f5.poolMemberExist(pool, membername) {
			remove := true
			splittedpool := strings.Split(pool.Name, "_")[0]
			for _, route := range routes {
				found := false
				if val, ok := route.Annotations[controller.CustomHostAnnotation]; ok {
					if val == f5.partition {
						found = true
					}
				}
				if (strings.HasSuffix(route.Spec.Host, hosttowatch) || found) && route.Spec.Host == splittedpool {
					remove = false
					break
				}
			}
			if remove {
				hosts[splittedpool] = true
			}
		}
	}
	return hosts
}

// PreUpdate checks are we running in HA mode, if yes write to active member
func (f5 *ProviderF5) PreUpdate() {
	// skip if no HA turned on
	if len(f5.addresses) == 1 {
		return
	}
	device, err := f5.session.GetCurrentDevice()
	if err != nil {
		msg := fmt.Sprintf("Error in PreUpdate %v", err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"stage": "preupdate"})
		}
		log.Printf(msg)
		return
	}
	if device.FailoverState == "standby" {
		log.Printf("changing f5.session to active member")
		count := f5.currentaddr + 1
		if len(f5.addresses) <= count {
			count = 0
		}
		f5.currentaddr = count
		f5.session = bigip.NewSession(f5.addresses[f5.currentaddr], f5.username, f5.password, nil)
	}
}

// PostUpdate syncs the configuration in f5 cluster
func (f5 *ProviderF5) PostUpdate() {
	// skip if no HA turned on
	if len(f5.addresses) == 1 {
		return
	}
	err := f5.session.ConfigSyncToGroup(f5.groupname)
	if err != nil {
		msg := fmt.Sprintf("Error in PostUpdate %v", err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"stage": "postupdate"})
		}
		log.Printf(msg)
	}
}

// Calls returns list of methodcalls, not in use in this provider
func (f5 *ProviderF5) Calls() []string {
	return nil
}

// CleanCalls cleans calls, not in use in this provider
func (f5 *ProviderF5) CleanCalls() {
}
