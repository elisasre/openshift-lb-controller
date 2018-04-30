/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package f5

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ElisaOyj/openshift-lb-controller/pkg/controller"
	v1 "github.com/openshift/api/route/v1"
	bigip "github.com/scottdware/go-bigip"
)

const providerName = "f5"

// ProviderF5 is an implementation of Interface for F5.
type ProviderF5 struct {
	session           *bigip.BigIP
	Clusteralias      string
	ClusterPRIOnumber int
	username          string
	password          string
	addresses         []string
	currentaddr       int
	groupname         string
}

func init() {
	controller.RegisterProvider(providerName, newProviderF5())
}

func newProviderF5() *ProviderF5 {
	f5 := ProviderF5{
		ClusterPRIOnumber: 0,
		currentaddr:       0,
		groupname:         "cluster",
	}
	return &f5
}

func alreadyExist(err error) bool {
	if strings.HasSuffix(err.Error(), "already exists in partition Common.") {
		return true
	}
	return false
}

// Initialize initilizes new provider
func (f5 *ProviderF5) Initialize() {
	address := os.Getenv("F5_ADDR")
	if len(address) == 0 {
		panic("F5_ADDR environment variable needed")
	}

	groupname := os.Getenv("F5_CLUSTERGROUP")
	if len(groupname) != 0 {
		f5.groupname = groupname
	}
	f5.addresses = strings.Split(address, ",")

	username := os.Getenv("F5_USER")
	if len(username) == 0 {
		panic("F5_USER environment variable needed")
	}
	password := os.Getenv("F5_PASSWORD")
	if len(password) == 0 {
		panic("F5_PASSWORD environment variable needed")
	}
	f5.username = username
	f5.password = password
	f5.session = bigip.NewSession(f5.addresses[0], f5.username, f5.password, nil)
	number := os.Getenv("CLUSTER_PRIO")
	if len(number) == 0 {
		panic("CLUSTER_PRIO environment variable needed")
	}
	i, err := strconv.Atoi(number)
	if err != nil {
		log.Println(err)
	} else {
		f5.ClusterPRIOnumber = i
	}
}

// CreatePool creates new loadbalancer pool
func (f5 *ProviderF5) CreatePool(name string, port string) error {
	err := f5.session.CreatePool(name + "_" + port)
	if err != nil {
		if !alreadyExist(err) {
			return err
		}
	}
	return nil
}

// AddPoolMember adds new member to pool
func (f5 *ProviderF5) AddPoolMember(membername string, name string, port string) error {
	f5.Clusteralias = membername
	err := f5.session.AddPoolMember(name+"_"+port, membername+":"+port)
	if err != nil {
		if !alreadyExist(err) {
			return err
		}
	}
	return nil
}

func (f5 *ProviderF5) modifyMember(name string, port string) {
	// we need use this because getpoolmember is not working correctly
	members, err := f5.session.PoolMembers(name + "_" + port)
	if err != nil {
		log.Printf("error in getpoolmembers %v", err)
		return
	}
	for _, item := range members.PoolMembers {
		if item.Name == f5.Clusteralias+":"+port {
			config := &bigip.PoolMember{
				FullPath:      item.FullPath,
				PriorityGroup: f5.ClusterPRIOnumber,
			}
			err = f5.session.PatchPoolMember(name+"_"+port, config)
			if err != nil {
				log.Printf("error in modifyMember %v", err)
			}
			break
		}
	}
}

// ModifyPool modifies loadbalancer pool
func (f5 *ProviderF5) ModifyPool(name string, port string, loadBalancingMethod string, pga int) error {
	pool, err := f5.session.GetPool(name + "_" + port)
	if err != nil {
		return err
	}
	targetmode := loadBalancingMethod
	if len(loadBalancingMethod) == 0 {
		targetmode = "round-robin"
	}
	log.Printf("changing pool %s loadbalancingmode to %s", name+"_"+port, targetmode)
	pool.LoadBalancingMode = targetmode
	log.Printf("changing pool %s pga to %d", name+"_"+port, pga)
	pool.MinActiveMembers = pga
	f5.modifyMember(name, port)
	err = f5.session.ModifyPool(name+"_"+port, pool)
	if err != nil {
		return err
	}
	return nil
}

// CreateMonitor creates new monitor
func (f5 *ProviderF5) CreateMonitor(name string, port string, host string, uri string, httpMethod string, interval int, timeout int) error {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	err := f5.session.CreateMonitor(name+"_"+port, scheme, interval, timeout, httpMethod+" "+uri+" HTTP/1.1\r\nHost:"+host+"  \r\nConnection: Close\r\n\r\n", "^HTTP.1.(0|1) ([2|3]0[0-9])", scheme)
	if err != nil {
		if !alreadyExist(err) {
			return err
		}
	}
	return nil
}

// ModifyMonitor modifies monitor
func (f5 *ProviderF5) ModifyMonitor(name string, port string, host string, uri string, httpMethod string, interval int, timeout int) error {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	config := &bigip.Monitor{
		Interval:   interval,
		Timeout:    timeout,
		SendString: httpMethod + " " + uri + " HTTP/1.1\r\nHost:" + host + "  \r\nConnection: Close\r\n\r\n",
	}
	err := f5.session.PatchMonitor(name+"_"+port, scheme, config)
	if err != nil {
		return err
	}
	return nil
}

// AddMonitorToPool adds monitor to pool
func (f5 *ProviderF5) AddMonitorToPool(name string, port string) error {
	err := f5.session.AddMonitorToPool(name+"_"+port, name+"_"+port)
	if err != nil {
		if !alreadyExist(err) {
			return err
		}
	}
	return nil
}

// DeletePoolMember delete pool member
func (f5 *ProviderF5) DeletePoolMember(membername string, poolname string, poolport string) error {
	return f5.session.DeletePoolMember(poolname+"_"+poolport, membername+":"+poolport)
}

// CheckAndClean checks pool members and if 0 members left in pool, delete monitor and delete pool
func (f5 *ProviderF5) CheckAndClean(name string, port string) {
	scheme := "http"
	if port == "443" {
		scheme = "https"
	}
	members, err := f5.session.PoolMembers(name + "_" + port)
	if err != nil {
		log.Printf("error retrieving poolmembers %s %v", name+"_"+port, err)
	}
	if len(members.PoolMembers) == 0 {
		err = f5.session.DeletePool(name + "_" + port)
		if err != nil {
			log.Printf("error delete pool %s %v", name+"_"+port, err)
		}

		err = f5.session.DeleteMonitor(name+"_"+port, scheme)
		if err != nil {
			log.Printf("error delete monitor %s %v", name+"_"+port, err)
		}
	}
}

func (f5 *ProviderF5) poolMemberExist(pool bigip.Pool, membername string) bool {
	members, err := f5.session.PoolMembers(pool.Name)
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

// CheckPools compares current load balancer setup and what routes we have. It returns list of pools which should be removed
func (f5 *ProviderF5) CheckPools(routes []v1.Route, hosttowatch string, membername string) []string {
	hosts := []string{}
	pools, err := f5.session.Pools()
	if err != nil {
		log.Printf("error fetching pool %v", err)
		return hosts
	}
	for _, pool := range pools.Pools {
		if f5.poolMemberExist(pool, membername) {
			remove := true
			splittedpool := strings.Split(pool.Name, "_")[0]
			for _, route := range routes {
				splittedroute := strings.Split(route.Spec.Host, ".")[0]
				if strings.HasSuffix(route.Spec.Host, hosttowatch) && splittedroute == splittedpool {
					remove = false
					break
				}
			}
			if remove {
				hosts = append(hosts, splittedpool+".")
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
		log.Printf("error in PreUpdate %v", err)
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
		log.Printf("error in PostUpdate %v", err)
	}
}

// Calls returns list of methodcalls, not in use in this provider
func (f5 *ProviderF5) Calls() []string {
	return nil
}

// CleanCalls cleans calls, not in use in this provider
func (f5 *ProviderF5) CleanCalls() {
}
