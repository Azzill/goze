/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package balancer

import (
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/discover"
)

type Balancer interface {
	//pick the best instance for a request
	PickInstance(serviceName string, ctx *common.RequestCtx) *BalancedService

	//notify the balancer good call or a network problem
	NotifyEffect(service *BalancedService, ok bool)

	//refresh balancer's instance
	RefreshInstance(instanceList map[string][]discover.MicroService)
}

type BalancedService struct {
	Client   *discover.RestClient   //rest client for rpc
	Service  *discover.MicroService //picked instance
	balancer Balancer               //used for notifying
	index    int                    //used for weighted cluster
}
