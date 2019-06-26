/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package discover

import (
	"errors"
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/server"
	"github.com/azzill/goze/util"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type RegisterServer struct {
	sync.Mutex
	ServiceInstances
	UpToDownDuration     time.Duration
	DownToRemoveDuration time.Duration
	heartbeatTicker      *time.Ticker
}

func init() {
	//SEED for guid
	rand.Seed(time.Now().UnixNano())
}

func StartDiscoverServer(addr string, udd time.Duration, drd time.Duration) {
	registerServer := RegisterServer{sync.Mutex{},
		ServiceInstances{Instance: map[string][]MicroService{}, Version: 0},
		drd, udd, time.NewTicker(time.Second * 3)}
	defer registerServer.heartbeatTicker.Stop()
	//stop the ticker when server stop
	go func() {

		//heartbeat check
		for range registerServer.heartbeatTicker.C {
			registerServer.checkHeartbeat()
		}
	}()
	registerServer.start(addr)
}

func (s *RegisterServer) start(addr string) {
	ds := server.NewRestServer(addr, &server.HttpConfig{})

	// fetch list, heartbeat
	ds.Mapping(server.Get, "/_ds", func(ctx *common.RequestCtx) (i interface{}) {
		info := ServiceInfo{}
		version, err := strconv.ParseUint(ctx.QueryString["version"][0], 10, 64)
		if err != nil {
			return err
		}

		if err = ctx.ParseBody(&info); err != nil {
			return err
		}

		//update last heartbeat
		validId := false
		for _, ins := range s.Instance[info.ServiceName] {
			if ins.InstanceId == info.Guid {
				ins.LastHeartbeat = time.Now()
				ins.Status = Up
				validId = true
				break
			}
		}

		if !validId {
			return errors.New("NoSuchInstance")
		}

		if version != s.Version {
			return s.filteredInstance()
		}

		return nil
	})

	// register
	ds.Mapping(server.Post, "/_ds", func(ctx *common.RequestCtx) (i interface{}) {
		ri := RegisterRequest{}
		e := ctx.ParseBody(&ri)
		if e != nil {
			return e
		}
		ms := MicroService{}
		e = util.Map(&ri, &ms)
		if e != nil {
			return e
		}

		ms.InstanceId = GUID(rand.Uint64())
		ms.RegisterTime = time.Now()
		ms.LastHeartbeat = ms.RegisterTime
		ms.Status = Up
		ms.Address = util.ParseIPAddr(ctx.Request.RemoteAddr)

		s.Lock()
		s.Instance[ms.ServiceName] = append(s.Instance[ms.ServiceName], ms)
		s.Version++
		s.Unlock()

		return ms
	})

	// offline
	ds.Mapping(server.Delete, "/_ds", func(ctx *common.RequestCtx) (i interface{}) {
		ur := ServiceInfo{}

		if e := ctx.ParseBody(&ur); e != nil {
			return e
		}
		ins := s.Instance[ur.ServiceName]
		for i := range ins {
			// only unregister from original ip address
			if ins[i].InstanceId == ur.Guid && ins[i].Address == util.ParseIPAddr(ctx.Request.RemoteAddr) {
				if len(ins) == 1 {
					delete(s.Instance, ur.ServiceName)
				} else if i == len(ins)-1 {
					s.Lock()
					s.Instance[ur.ServiceName] = ins[:i-1]
					s.Version++
					s.Unlock() //forgive the bullshit codes
				} else if i == 0 {
					s.Lock()
					s.Instance[ur.ServiceName] = ins[i+1:]
					s.Version++
					s.Unlock()
				} else {
					s.Lock()
					s.Instance[ur.ServiceName] = append(ins[:i-1], ins[i+1:]...)
					s.Version++
					s.Unlock()
				}
				return nil
			}
		}
		return errors.New("NoSuchInstance")
	})

	ds.StartServer()
}

func (s *RegisterServer) filteredInstance() *RegisterServer {
	s.Lock()
	defer s.Unlock()
	//in case for concurrent update

	n := &RegisterServer{sync.Mutex{},
		ServiceInstances{Instance: map[string][]MicroService{}, Version: s.Version},
		s.UpToDownDuration, s.DownToRemoveDuration, s.heartbeatTicker}
	for key, instances := range s.Instance {
		services := make([]MicroService, 0, len(instances))
		for _, instance := range instances {

			//filter service with DOWN status
			if instance.Status == Down {
				continue
			}
			services = append(services, instance)
		}
		n.Instance[key] = services
	}

	return n
}

func (s *RegisterServer) checkHeartbeat() {
	s.Lock()
	defer s.Unlock()
	//in case for concurrent update

	for key, instances := range s.Instance {
		for i, instance := range instances {

			//filter service with DOWN status
			if instance.Status == Up {
				if time.Now().Sub(instance.LastHeartbeat) > s.UpToDownDuration {
					s.Instance[key][i].Status = Down
					logger.Warn("Change instance status to DOWN for no heartbeat -",
						instance.ServiceName, instance.InstanceId)
				}
			}

			//remove from instance list
			if instance.Status == Down {
				if time.Now().Sub(instance.LastHeartbeat) > (s.UpToDownDuration + s.DownToRemoveDuration) {

					if len(instances) == 1 {
						s.Instance[key] = []MicroService{}
						continue
					}

					services := make([]MicroService, 0, len(instances)-1)
					for i2, service := range s.Instance[key] {
						if i2 != i {
							services = append(services, service)
						}
					}
					s.Instance[key] = services

					logger.Warn("Remove instance status DOWN with no heartbeat -",
						instance.ServiceName, instance.InstanceId)
				}
			}

		}
	}

}
