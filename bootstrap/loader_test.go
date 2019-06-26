/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package bootstrap

import (
	"github.com/azzill/goze/common"
	"github.com/azzill/goze/server"
	"testing"
)

type Service struct {
}

func (*Service) SayHello() string {
	return "Hello"
}

type Controller struct {
	Service *Service `inject:"true"`
}

func (c *Controller) Mapping(s *server.RestServer) {
	s.GET("/", func(ctx *common.RequestCtx) (i interface{}) {
		return "Hello"
	})
}

func TestStartGozeApplication(t *testing.T) {
	StartGozeApplication(&Service{}, &Controller{})
}
