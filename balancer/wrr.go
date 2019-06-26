/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package balancer

import (
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/discover"
	"sync"
	"time"
)

type weightedCluster struct {
	sync.Mutex
	instances       []discover.MicroService
	currentWeight   []int64
	effectiveWeight []uint
}

type WRRBalancer struct {
	Balancer

	sync.RWMutex
	timeout  time.Duration
	services map[string]*weightedCluster
}

func NewWRRBalancer(timeout time.Duration) *WRRBalancer {
	return &WRRBalancer{timeout: timeout, services: map[string]*weightedCluster{}}
}

func newWeightedCluster(instances []discover.MicroService) *weightedCluster {
	clusters := &weightedCluster{instances: instances, currentWeight: make([]int64, len(instances))}
	effectiveWeight := make([]uint, len(instances))
	for i, ins := range instances {
		effectiveWeight[i] = ins.Weight
	}
	clusters.effectiveWeight = effectiveWeight
	return clusters
}

func (c *weightedCluster) pickInstance() (service *discover.MicroService, index int) {
	maxI := 0
	var totalWeight int64 = 0
	maxWeight := c.currentWeight[0] + int64(c.effectiveWeight[0])
	for i := range c.instances {
		c.currentWeight[i] = c.currentWeight[i] + int64(c.effectiveWeight[i])
		totalWeight += c.currentWeight[i]
		if c.currentWeight[i] > maxWeight {
			maxI = i
			maxWeight = c.currentWeight[i]
		}

	}
	c.currentWeight[maxI] -= totalWeight
	return &c.instances[maxI], maxI
}

func (s *WRRBalancer) PickInstance(serviceName string, ctx *common.RequestCtx) *BalancedService {
	s.RLock()
	defer s.RUnlock()
	//lock in case concurrency update and reading

	cluster := s.services[serviceName]
	if cluster == nil {
		return nil
	} else {
		service, index := cluster.pickInstance()
		return &BalancedService{
			Client:   discover.NewRestClient(s.timeout, service),
			Service:  service,
			balancer: s,
			index:    index,
		}
	}

}

func (s *WRRBalancer) NotifyEffect(bs *BalancedService, ok bool) {
	cluster := s.services[bs.Service.ServiceName]

	if ok && cluster.effectiveWeight[bs.index] < cluster.instances[bs.index].Weight {
		cluster.Lock()
		defer cluster.Unlock()

		//lock in case concurrency write
		if cluster.effectiveWeight[bs.index] < cluster.instances[bs.index].Weight {
			cluster.effectiveWeight[bs.index]++
			return
		}
	}

	if !ok && cluster.effectiveWeight[bs.index] > 0 {
		cluster.Lock()
		defer cluster.Unlock()

		//lock in case concurrency write
		if cluster.effectiveWeight[bs.index] > 0 {
			cluster.effectiveWeight[bs.index]--
			return
		}
	}
}

func (s *WRRBalancer) RefreshInstance(instanceList map[string][]discover.MicroService) {
	s.Lock()
	defer s.Unlock()
	//lock in case concurrent update and read

	s.services = map[string]*weightedCluster{}

	for k, instances := range instanceList {
		cluster := &weightedCluster{
			effectiveWeight: make([]uint, len(instances)),
			currentWeight:   make([]int64, len(instances)),
		}
		s.services[k] = cluster
		cluster.instances = instances

		//initialize all instances' effective weight
		for i := range instances {
			cluster.effectiveWeight[i] = instances[i].Weight
		}
	}
}
