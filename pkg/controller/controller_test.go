/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package controller

import (
	fake "github.com/ElisaOyj/openshift-lb-controller/pkg/controller/providers/fakeprovider"
	v1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreate(t *testing.T) {
	fakeRouteController := &RouteController{}
	fakeRouteController.hosttowatch = "test.com"
	fakeRouteController.clusteralias = "dc1"

	newfake := fake.NewFakeProvider()
	fakeRouteController.provider = ProviderInterface(newfake)

	obj := &v1.Route{
		Spec: v1.RouteSpec{
			Host: "foo.test.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.createRoute(obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted createpool")
	}
	fakeRouteController.provider.CleanCalls()
	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted clean calls")
	}

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				CustomHostAnnotation: "enabled",
			},
		},
		Spec: v1.RouteSpec{
			Host: "leet.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.createRoute(obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted createpool")
	}
	fakeRouteController.provider.CleanCalls()
	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted clean calls")
	}

	obj = &v1.Route{
		Spec: v1.RouteSpec{
			Host: "foo.testx.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.createRoute(obj)

	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted skip")
	}
}

func TestDelete(t *testing.T) {
	fakeRouteController := &RouteController{}
	fakeRouteController.hosttowatch = "test.com"
	fakeRouteController.clusteralias = "dc1"

	newfake := fake.NewFakeProvider()
	fakeRouteController.provider = ProviderInterface(newfake)

	obj := &v1.Route{
		Spec: v1.RouteSpec{
			Host: "foo.test.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.deleteRoute(obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted deletepool")
	}
	fakeRouteController.provider.CleanCalls()
	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted clean calls")
	}

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				CustomHostAnnotation: "enabled",
			},
		},
		Spec: v1.RouteSpec{
			Host: "leet.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.deleteRoute(obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted deletepool")
	}
	fakeRouteController.provider.CleanCalls()
	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted clean calls")
	}

	obj = &v1.Route{
		Spec: v1.RouteSpec{
			Host: "foo.testx.com",
			To:   v1.RouteTargetReference{Name: "other"},
			TLS:  &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{{}},
		},
	}

	fakeRouteController.deleteRoute(obj)

	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted skip")
	}
}

func TestUpdate(t *testing.T) {
	fakeRouteController := &RouteController{}
	fakeRouteController.hosttowatch = "test.com"
	fakeRouteController.clusteralias = "dc1"

	newfake := fake.NewFakeProvider()
	fakeRouteController.provider = ProviderInterface(newfake)

	obj := &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 := &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	obj = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.texst.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.texst.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted skip")
	}

	fakeRouteController.provider.CleanCalls()
	if len(fakeRouteController.provider.Calls()) != 0 {
		t.Errorf("excepted clean calls")
	}

	// delete
	obj = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.texst.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted delete")
	}
	if fakeRouteController.provider.Calls()[1] != "DeletePoolMember" {
		t.Errorf("excepted delete")
	}
	fakeRouteController.provider.CleanCalls()

	// create
	obj = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.tesxt.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted create")
	}
	if fakeRouteController.provider.Calls()[1] != "CreatePool" {
		t.Errorf("excepted create")
	}
	fakeRouteController.provider.CleanCalls()

	// create
	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				CustomHostAnnotation: "enabled",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj2, obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted create")
	}
	if len(fakeRouteController.provider.Calls()) > 1 && fakeRouteController.provider.Calls()[1] != "CreatePool" {
		t.Errorf("excepted create")
	}
	fakeRouteController.provider.CleanCalls()

	// delete
	obj = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foobar.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				CustomHostAnnotation: "enabled",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foobar.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj2, obj)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted delete")
	}
	if len(fakeRouteController.provider.Calls()) > 1 && fakeRouteController.provider.Calls()[1] != "DeletePoolMember" {
		t.Errorf("excepted delete")
	}
	fakeRouteController.provider.CleanCalls()

}

func TestUpdateAnnotation(t *testing.T) {
	fakeRouteController := &RouteController{}
	fakeRouteController.hosttowatch = "test.com"
	fakeRouteController.clusteralias = "dc1"

	newfake := fake.NewFakeProvider()
	fakeRouteController.provider = ProviderInterface(newfake)

	obj := &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				healthCheckPathAnnotation: "test",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 := &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update healthcheckpath")
	}

	fakeRouteController.provider.CleanCalls()
	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update healthcheckpath")
	}
	fakeRouteController.provider.CleanCalls()

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				healthCheckMethodAnnotation: "test",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to default healthcheckmethod")
	}

	fakeRouteController.provider.CleanCalls()
	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update healthcheckmethod")
	}
	fakeRouteController.provider.CleanCalls()

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				poolRouteMethodAnnotation: "test",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to default poolroutemethod")
	}
	fakeRouteController.provider.CleanCalls()

	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update poolroutemethod")
	}
	fakeRouteController.provider.CleanCalls()

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				poolRouteMethodAnnotation: "test",
				CustomHostAnnotation:      "enabled",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				CustomHostAnnotation: "enabled",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to default poolroutemethod")
	}
	fakeRouteController.provider.CleanCalls()

	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update poolroutemethod")
	}
	fakeRouteController.provider.CleanCalls()

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				poolPGARouteMethodAnnotation: "1",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to default update pga")
	}
	fakeRouteController.provider.CleanCalls()

	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update pga")
	}

	fakeRouteController.provider.CleanCalls()

	obj = &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				maintenanceAnnotation: "",
			},
		},
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	obj2 = &v1.Route{
		Spec: v1.RouteSpec{
			To:  v1.RouteTargetReference{Name: "other"},
			TLS: &v1.TLSConfig{},
		},
		Status: v1.RouteStatus{
			Ingress: []v1.RouteIngress{
				{
					Host: "foo.test.com",
				},
			},
		},
	}

	fakeRouteController.updateRoute(obj, obj2)

	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to default update maintenance")
	}
	fakeRouteController.provider.CleanCalls()

	fakeRouteController.updateRoute(obj2, obj)
	if len(fakeRouteController.provider.Calls()) == 0 {
		t.Errorf("excepted to update maintenance")
	}
}
