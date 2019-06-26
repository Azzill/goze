/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package balancer

import (
	"github.com/azzill/goze/discover"
	"testing"
	"time"
)

func TestWRRBalancer(t *testing.T) {
	balancer := NewWRRBalancer(time.Second * 10)
	s1 := []discover.MicroService{
		*discover.NewWeightedMicroService("SN1", 1001, 1),
		*discover.NewWeightedMicroService("SN1", 1002, 2),
		*discover.NewWeightedMicroService("SN1", 1003, 3),
	}

	s2 := []discover.MicroService{
		*discover.NewWeightedMicroService("SN2", 1001, 1),
		*discover.NewWeightedMicroService("SN2", 1002, 2),
		*discover.NewWeightedMicroService("SN2", 1003, 3),
	}

	balancer.RefreshInstance(map[string][]discover.MicroService{
		"SN1": s1,
		"SN2": s2,
	})

	total := []int{0, 0, 0}
	for i := 0; i < 12; i++ {
		instance := balancer.PickInstance("SN1", nil)

		for i, s := range balancer.services["SN1"].currentWeight {
			println("current weight: ", i, s)
		}

		for i, s := range balancer.services["SN1"].effectiveWeight {
			println("effective weight: ", i, s)
		}

		total[instance.index]++
		println("INSTANCE:", instance.index)
	}

	if !(total[0] == 2 && total[1] == 4 && total[2] == 6) {
		t.Error("WRR not correct")
	}
}
