/*
Copyright (C) 2018 Elisa Oyj

SPDX-License-Identifier: Apache-2.0
*/

package f5

import (
	bigip "github.com/scottdware/go-bigip"
	"testing"
)

func TestCreate(t *testing.T) {

	newf5 := NewProviderF5()
	newf5.Clusteralias = "xx"
	newf5.addresses = []string{"xxx"}
	newf5.username = "yy"
	newf5.password = "xx"
	newf5.partition = "xx"
	newf5.session = bigip.NewSession(newf5.addresses[0], newf5.username, newf5.password, nil)

	newf5.PreUpdate()

	err := newf5.CreatePool("test", "80")
	if err != nil {
		t.Errorf("%v", err)

	}
	err = newf5.AddPoolMember(newf5.Clusteralias, "test", "80")
	if err != nil {
		t.Errorf("%v", err)

	}

	err = newf5.ModifyPool("test", "80", "", 1, false, 1)
	if err != nil {
		t.Errorf("%v", err)
	}

	err = newf5.CreateMonitor("test", "80", "foobar.com", "http", 3, 10)
	if err != nil {
		t.Errorf("%v", err)
	}

	err = newf5.AddMonitorToPool("test", "80")
	if err != nil {
		t.Errorf("%v", err)
	}
	pools, err := newf5.getPools()
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(pools.Pools) != 1 {
		t.Errorf("should be one pool")

	}

	err = newf5.DeletePoolMember(newf5.Clusteralias, "test", "80")
	if err != nil {
		t.Errorf("%v", err)
	}
	
	newf5.CheckAndClean("test", "80")

	pools, err = newf5.getPools()
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(pools.Pools) != 0 {
		t.Errorf("should be zero pool")

	}
	newf5.PostUpdate()


}
