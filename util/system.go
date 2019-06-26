/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package util

type SystemLoad struct {
	CPUUTotal   uint8 `json:"cpu_total"`
	CPUUsed     uint8 `json:"cpu_use"`
	MemoryTotal uint  `json:"memory_total"`
	MemoryUsed  uint  `json:"memory_used"`
}

func AcquireSystemLoad() SystemLoad {
	return SystemLoad{}
}
