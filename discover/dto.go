/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package discover

import (
	"github.com/azzill/goze/log"
	"time"
)

//This file defines all request DTOs used for discover server and clients

//logger
var logger = log.NewLogger("Discover")

type RegisterRequest struct {
	ServiceName string `json:"service_name"`
	Weight      uint   `json:"weight"`
	Port        uint   `json:"port"`
	CPUUse      uint   `json:"cpu_use"`
	MemoryTotal uint   `json:"memory_total"`
	MemoryUsed  uint   `json:"memory_used"`
}

type ServiceInfo struct {
	ServiceName string `json:"service_name"`
	Guid        GUID   `json:"guid"`
}

type ServiceInstances struct {
	Instance map[string][]MicroService `json:"instance"`
	Version  uint64                    `json:"version"`
}

type RegisteredInfo struct {
	InstanceId   GUID          `json:"instance_id"`
	RegisterTime time.Time     `json:"register_time"`
	Status       ServiceStatus `json:"status"`
}

type HeartbeatInfo struct {
	CPUUse      uint `json:"cpu_use"`
	MemoryTotal uint `json:"memory_total"`
	MemoryUsed  uint `json:"memory_used"`
}
