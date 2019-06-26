/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package midware

import (
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/log"
	"sort"
	"sync"
)

type InterceptorChain struct {
	sync.Mutex
	interceptor []Interceptor
}

var logger = log.NewLogger("Interceptor")

func NewInterceptorChain() *InterceptorChain {
	return &InterceptorChain{interceptor: []Interceptor{}}
}

func (r *InterceptorChain) AddInterceptor(interceptor Interceptor) {
	r.Lock()
	ins := append(r.interceptor, interceptor)
	sort.SliceStable(ins, func(i, j int) bool {
		return ins[i].Priority() < ins[j].Priority()
	})
	r.interceptor = ins
	r.Unlock()
}

// interceptor chain returns a bool tell if the request should be blocked or resumed
// second parameter will be treat as the response body
func (r *InterceptorChain) CallInterceptors(ctx *common.RequestCtx) (bool, interface{}) {
	for _, i := range r.interceptor {
		action, ret := i.Intercept(ctx)
		switch action {
		case Continue:
			if ret != nil {
				logger.Warn("Continued interceptors returns no nil value makes no sense")
			}
			continue
		case Block:
			return true, ret
		case Skip:
			if ret != nil {
				logger.Warn("Skipped interceptors returns no nil value makes no sense")
			}
			return false, nil
		}
	}
	return false, nil
}
