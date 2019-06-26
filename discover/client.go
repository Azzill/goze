/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package discover

import (
	"github.com/azzill/goze/server"
	"github.com/azzill/goze/util"
	"math"
	"time"
)

type GUID uint64
type ServiceStatus int

const logSender = "MicroService DC"

var heartbeatTicker *time.Ticker

const (
	_ ServiceStatus = iota
	Up
	Exhausted
	Down
)

//The instance of current micro-service, nil if not registered
type InstanceManager struct {
	current *MicroService
	info    ServiceInstances
	addr    string
}

type MicroService struct {
	InstanceId    GUID          `json:"instance_id"`
	RegisterTime  time.Time     `json:"register_time"`
	LastHeartbeat time.Time     `json:"last_heartbeat"`
	Status        ServiceStatus `json:"status"`

	Address     string `json:"address"`
	Port        uint   `json:"port"`
	ServiceName string `json:"service_name"`
	Weight      uint   `json:"weight"`

	CPUUse      uint `json:"cpu_use"`
	MemoryTotal uint `json:"memory_total"`
	MemoryUsed  uint `json:"memory_used"`
}

func NewWeightedMicroService(serviceName string, port uint, serviceWeigh uint) *MicroService {
	return &MicroService{ServiceName: serviceName, Port: port, Weight: serviceWeigh}
}

func NewMicroService(serviceName string, port uint) *MicroService {
	return &MicroService{ServiceName: serviceName, Port: port, Weight: math.MaxUint8}
}

func (InstanceManager *InstanceManager) Register(service *MicroService, remoteAddr string, heartbeatDuration time.Duration) {

	resp := RegisteredInfo{}
	_, e := util.RestRequest(server.Post, remoteAddr, RegisterRequest{Weight: service.Weight, ServiceName: service.ServiceName, Port: service.Port}, &resp)

	if e != nil {
		logger.Error(e)
	}

	if InstanceManager.current == nil {
		InstanceManager.current = service
	}

	e = util.Map(&resp, InstanceManager.current)

	if e != nil {
		InstanceManager.current = nil
		logger.Error(e)
	}

	InstanceManager.addr = remoteAddr
	InstanceManager.info = ServiceInstances{}

	heartbeatTicker = time.NewTicker(heartbeatDuration)
	go func() {
		for range heartbeatTicker.C {
			InstanceManager.FetchInstanceInfo()
		}
	}()

	logger.Info("Instance has been registered as", resp.InstanceId)
	InstanceManager.FetchInstanceInfo()
}

//unregister service when exit
func (InstanceManager *InstanceManager) Unregister() {
	heartbeatTicker.Stop()
	heartbeatTicker = nil
	InstanceManager.addr = ""
	InstanceManager.current = nil
	InstanceManager.info = ServiceInstances{}

	if InstanceManager.current == nil {
		return
	}

	_, e := util.RestRequest(server.Delete, InstanceManager.addr,
		&ServiceInfo{Guid: InstanceManager.current.InstanceId, ServiceName: InstanceManager.current.ServiceName}, nil)

	if e != nil {
		logger.Error(e)
	}

}

func (InstanceManager *InstanceManager) FetchInstanceInfo() {
	logger.Info("Fetching instance version")
	instances := ServiceInstances{}
	b, e := util.RestRequest(server.Put, InstanceManager.addr, &ServiceInfo{Guid: InstanceManager.current.InstanceId,
		ServiceName: InstanceManager.current.ServiceName}, &instances)
	if e != nil {
		logger.Error("Failed to fetch instances from register server", InstanceManager.addr, e)
		return
	}

	if b {
		InstanceManager.info = instances
		logger.Info("Instance list updated, new version", InstanceManager.info.Version)
	} else {
		logger.Info("Instance list is up to date")
	}
}
