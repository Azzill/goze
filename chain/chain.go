/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package chain

import (
	"github.com/azzill/goze/discover"
	"time"
)

type InvokeChainContext struct {
	ChainId uint32
	Stages  []StageContext
}

type StageContext struct {
	Start     time.Time
	End       time.Time
	Error     string
	Code      uint32
	ServiceId discover.GUID
}

type Tracer interface {
	PreviousInvoke() *StageContext
	IsInternalInvoke() bool
	hasError() bool
	CalculateTime() time.Duration
}
