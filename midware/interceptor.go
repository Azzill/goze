package midware

import (
	"github.com/azzill/goze/common"
)

type InterceptorAction int

const (
	_ InterceptorAction = iota

	// call next interceptor Routine
	// returned value must be nil
	Continue

	// block the request,
	// returned value can be assigned by the second parameter
	Block

	// skip other interceptors, enter normal handler, not recommended
	// returned value must be nil
	Skip
)

// eg: return Continue, nil
//	   return Block, "No Permission"
//	   return Skip, nil

type Interceptor interface {

	// higher prior
	Priority() int
	Intercept(ctx *common.RequestCtx) (InterceptorAction, interface{})
}
