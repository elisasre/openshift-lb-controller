/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ElisaOyj/openshift-lb-controller/pkg/common"
	"github.com/getsentry/raven-go"
	v1r "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	healthCheckPathAnnotation    = "route.elisa.fi/path"
	healthCheckMethodAnnotation  = "route.elisa.fi/method"
	poolRouteMethodAnnotation    = "route.elisa.fi/lbmethod"
	poolPGARouteMethodAnnotation = "route.elisa.fi/poolpga"
	// CustomHostAnnotation is annotation which enables lb for custom route hostname
	CustomHostAnnotation = "route.elisa.fi/lbenabled"

	maintenanceAnnotation = "route.elisa.fi/maintenance"
)

// RouteController watches the kubernetes api for changes to routes
type RouteController struct {
	routeInformer cache.SharedIndexInformer
	kclient       *kubernetes.Clientset
	routeclient   *routev1.RouteV1Client
	hosttowatch   string
	clusteralias  string
	provider      ProviderInterface
}

// Run starts the process for listening for route changes and acting upon those changes.
func (c *RouteController) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)

	// Execute go function
	go c.routeInformer.Run(stopCh)

	// Wait till we receive a stop signal
	<-stopCh
}

// NewRouteController creates a new RouteController
func NewRouteController(kclient *kubernetes.Clientset, config *restclient.Config) *RouteController {
	routeWatcher := &RouteController{}

	routeV1Client, err := routev1.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	routeInformer := cache.NewSharedIndexInformer(

		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return routeV1Client.Routes(v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return routeV1Client.Routes(v1.NamespaceAll).Watch(options)
			},
		},
		&v1r.Route{},
		3*time.Minute,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	routeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    routeWatcher.createRoute,
		UpdateFunc: routeWatcher.updateRoute,
		DeleteFunc: routeWatcher.deleteRoute,
	})

	routeWatcher.kclient = kclient
	routeWatcher.routeclient = routeV1Client
	routeWatcher.routeInformer = routeInformer

	host := os.Getenv("SUFFIXHOST")
	if len(host) == 0 {
		err := errors.New("SUFFIXHOST environment variable needed")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}
	routeWatcher.hosttowatch = host
	clustername := os.Getenv("CLUSTERALIAS")
	if len(clustername) == 0 {
		err := errors.New("CLUSTERALIAS environment variable needed")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}
	routeWatcher.clusteralias = clustername
	provider := routeWatcher.InitProvider()
	if provider == nil {
		err := errors.New("Could not find working LB provider")
		if common.SentryEnabled() {
			raven.CaptureErrorAndWait(err, nil)
		}
		panic(err)
	}
	routeWatcher.provider = provider
	routeWatcher.provider.Initialize()
	routeWatcher.cleanUp()
	return routeWatcher
}

// cleanUp will be executed in start. It will compare LB and openshift configurations
// in case of openshift routes are deleted, the LB configurations needs to be deleted as well
func (c *RouteController) cleanUp() {

	routes, err := c.routeclient.Routes(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.Printf("error fetching routes %v", err)
		return
	}
	poolsToBeRemoved := c.provider.CheckPools(routes.Items, c.hosttowatch, c.clusteralias)
	for host := range poolsToBeRemoved {
		c.checkExternalLBDoesNotExists(host)
	}
}

func (c *RouteController) checkExternalLBDoesExists(host string, uri string, httpMethod string, loadBalancingMethod string, pga int, maintenance bool) {
	c.provider.PreUpdate()
	err := c.provider.CreatePool(host, "80")
	if err != nil {
		msg := fmt.Sprintf("Error in CreatePool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.CreatePool(host, "443")
	if err != nil {
		msg := fmt.Sprintf("Error in CreatePool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	err = c.provider.AddPoolMember(c.clusteralias, host, "80")
	if err != nil {
		msg := fmt.Sprintf("Error in AddPoolMember %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.AddPoolMember(c.clusteralias, host, "443")
	if err != nil {
		msg := fmt.Sprintf("Error in AddPoolMember %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	err = c.provider.ModifyPool(host, "80", loadBalancingMethod, pga, maintenance)
	if err != nil {
		msg := fmt.Sprintf("Error in ModifyPool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.ModifyPool(host, "443", loadBalancingMethod, pga, maintenance)
	if err != nil {
		msg := fmt.Sprintf("Error in ModifyPool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	err = c.provider.CreateMonitor(host, "80", uri, httpMethod, 3, 10)
	if err != nil {
		msg := fmt.Sprintf("Error in CreateMonitor %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.CreateMonitor(host, "443", uri, httpMethod, 3, 10)
	if err != nil {
		msg := fmt.Sprintf("Error in CreateMonitor %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	err = c.provider.AddMonitorToPool(host, "80")
	if err != nil {
		msg := fmt.Sprintf("Error in AddMonitorToPool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.AddMonitorToPool(host, "443")
	if err != nil {
		msg := fmt.Sprintf("Error in AddMonitorToPool %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	c.provider.PostUpdate()
	log.Printf("add external lb configuration host: %s to clusteralias: %s", host, c.clusteralias)
}

func (c *RouteController) checkExternalLBDoesNotExists(host string) {
	c.provider.PreUpdate()
	err := c.provider.DeletePoolMember(c.clusteralias, host, "80")
	if err != nil {
		msg := fmt.Sprintf("Error in DeletePoolMember %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}
	err = c.provider.DeletePoolMember(c.clusteralias, host, "443")
	if err != nil {
		msg := fmt.Sprintf("Error in DeletePoolMember %s: %v", host, err)
		if common.SentryEnabled() {
			raven.CaptureMessage(msg, map[string]string{"host": host})
		}
		log.Printf(msg)
	}

	// if 0 members left in pool, cleanup monitor and delete pool
	c.provider.CheckAndClean(host, "80")
	c.provider.CheckAndClean(host, "443")

	c.provider.PostUpdate()
	log.Printf("delete external lb configuration host: %s from clusteralias: %s", host, c.clusteralias)
}

func (c *RouteController) updateRoute(old interface{}, obj interface{}) {
	route := obj.(*v1r.Route)
	routeold := old.(*v1r.Route)

	_, found := route.Annotations[CustomHostAnnotation]
	_, foundold := routeold.Annotations[CustomHostAnnotation]

	if len(routeold.Status.Ingress) > 0 && len(route.Status.Ingress) > 0 {
		// if old did not have and now it has
		if (!strings.HasSuffix(routeold.Status.Ingress[0].Host, c.hosttowatch) && strings.HasSuffix(route.Status.Ingress[0].Host, c.hosttowatch)) || (!foundold && found) {
			// read healthcheck path
			healthCheckPath, healthCheckMethod, loadBalancingMethod, pga, maintenance := overrideWithAnnotation(route)
			c.checkExternalLBDoesExists(route.Status.Ingress[0].Host, healthCheckPath, healthCheckMethod, loadBalancingMethod, pga, maintenance)
			// if old have and now it does not have
		} else if (strings.HasSuffix(routeold.Status.Ingress[0].Host, c.hosttowatch) && !strings.HasSuffix(route.Status.Ingress[0].Host, c.hosttowatch)) || (!found && foundold) {
			c.checkExternalLBDoesNotExists(routeold.Status.Ingress[0].Host)
			// check annotation changes here
		} else if strings.HasSuffix(route.Status.Ingress[0].Host, c.hosttowatch) || found {
			healthCheckPathold, healthCheckMethodold, loadBalancingMethodold, pgaold, maintenanceold := overrideWithAnnotation(routeold)
			healthCheckPath, healthCheckMethod, loadBalancingMethod, pga, maintenance := overrideWithAnnotation(route)
			host := route.Status.Ingress[0].Host
			c.provider.PreUpdate()
			if loadBalancingMethodold != loadBalancingMethod || pgaold != pga || maintenanceold != maintenance {
				err := c.provider.ModifyPool(host, "80", loadBalancingMethod, pga, maintenance)
				if err != nil {
					msg := fmt.Sprintf("Error in ModifyPool %s: %v", host, err)
					if common.SentryEnabled() {
						raven.CaptureMessage(msg, map[string]string{"host": host})
					}
					log.Printf(msg)
				}
				err = c.provider.ModifyPool(host, "443", loadBalancingMethod, pga, maintenance)
				if err != nil {
					msg := fmt.Sprintf("Error in ModifyPool %s: %v", host, err)
					if common.SentryEnabled() {
						raven.CaptureMessage(msg, map[string]string{"host": host})
					}
					log.Printf(msg)
				}
			}
			if healthCheckPathold != healthCheckPath || healthCheckMethodold != healthCheckMethod {
				err := c.provider.ModifyMonitor(host, "80", healthCheckPath, healthCheckMethod, 3, 10)
				if err != nil {
					msg := fmt.Sprintf("Error in ModifyMonitor %s: %v", host, err)
					if common.SentryEnabled() {
						raven.CaptureMessage(msg, map[string]string{"host": host})
					}
					log.Printf(msg)
				}
				err = c.provider.ModifyMonitor(host, "443", healthCheckPath, healthCheckMethod, 3, 10)
				if err != nil {
					msg := fmt.Sprintf("Error in ModifyMonitor %s: %v", host, err)
					if common.SentryEnabled() {
						raven.CaptureMessage(msg, map[string]string{"host": host})
					}
					log.Printf(msg)
				}
			}
			c.provider.PostUpdate()
		}
	}
}

func (c *RouteController) deleteRoute(obj interface{}) {
	route := obj.(*v1r.Route)
	_, found := route.Annotations[CustomHostAnnotation]
	// has suffix what we are interested, skip others
	if strings.HasSuffix(route.Spec.Host, c.hosttowatch) || found {
		c.checkExternalLBDoesNotExists(route.Spec.Host)
	}
}
func (c *RouteController) createRoute(obj interface{}) {
	route := obj.(*v1r.Route)
	_, found := route.Annotations[CustomHostAnnotation]
	// has suffix what we are interested, skip others
	if strings.HasSuffix(route.Spec.Host, c.hosttowatch) || found {
		// read healthcheck path
		healthCheckPath, healthCheckMethod, loadBalancingMethod, pga, maintenance := overrideWithAnnotation(route)
		c.checkExternalLBDoesExists(route.Spec.Host, healthCheckPath, healthCheckMethod, loadBalancingMethod, pga, maintenance)
	}
}

func overrideWithAnnotation(route *v1r.Route) (string, string, string, int, bool) {
	path := "/"
	method := "GET"
	lbmethod := ""
	maintenance := false
	pga := 0
	if annotationValue, ok := route.Annotations[healthCheckPathAnnotation]; ok {
		path = annotationValue
	}
	if annotationValue, ok := route.Annotations[healthCheckMethodAnnotation]; ok {
		method = annotationValue
	}
	if annotationValue, ok := route.Annotations[poolRouteMethodAnnotation]; ok {
		lbmethod = annotationValue
	}
	if annotationValue, ok := route.Annotations[poolPGARouteMethodAnnotation]; ok {
		i, err := strconv.Atoi(annotationValue)
		if err != nil {
			if common.SentryEnabled() {
				raven.CaptureError(err, nil)
			}
			log.Println(err)
		} else {
			pga = i
		}
	}
	if _, ok := route.Annotations[maintenanceAnnotation]; ok {
		maintenance = true
	}
	return path, method, lbmethod, pga, maintenance
}
