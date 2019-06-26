/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package context

import (
	"fmt"
	"github.com/azzill/goze/config"
	"github.com/azzill/goze/log"
	"github.com/azzill/goze/midware"
	"github.com/azzill/goze/server"
	"github.com/azzill/goze/sql"
	"reflect"
)

// The context of goze application instance
type ApplicationContext struct {
	Components    map[string]interface{}
	Configuration *config.CommonConfiguration
}

type Controller interface {
	//implement this yourself
	Mapping(server *server.RestServer)
}

type Configer interface {
	Config(cfg *config.CommonConfiguration) interface{}
}

func (c *ApplicationContext) GetComponent(name string) interface{} {
	return c.Components[name]
}

func (c *ApplicationContext) With(component interface{}) *ApplicationContext {

	if configer, ok := component.(Configer); ok {
		component = configer.Config(c.Configuration)
	}
	if reflect.TypeOf(component).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Component `%s`(%s) must be a pointer!", reflect.TypeOf(component).Name(),
			reflect.TypeOf(component).Kind().String()))
	}
	name := reflect.TypeOf(component).Elem().Name()
	if c.Components[name] != nil {
		panic(fmt.Sprintf("Component `%s` already exists.", name))

	}
	c.Components[name] = component
	switch component.(type) {
	case Controller:
		component.(Controller).Mapping(c.Components["RestServer"].(*server.RestServer))

	case server.ResponseWrapper:
		c.Components["RestServer"].(*server.RestServer).AddResponseWrapper(component.(server.ResponseWrapper))

	case midware.Interceptor:
		c.Components["RestServer"].(*server.RestServer).AddInterceptor(component.(midware.Interceptor))

	case *sql.SQL:
		c.Components["RestServer"].(*server.RestServer).WithSQL(component.(*sql.SQL))
	}
	return c
}

func (c *ApplicationContext) Inject() {
	logger := log.NewLogger("Component Injector")
	components := c.Components
	for _, v := range components {
		cElement := reflect.ValueOf(v).Elem()
		for i := 0; i < cElement.NumField(); i++ {

			// if have a field with assignable value, inject it
			if inject := reflect.TypeOf(v).Elem().Field(i).Tag.Get("inject"); inject == "" {
				continue //No Inject
			} else {

				//value should be true
				if inject != "true" {
					logger.Warn("Component:", cElement.Type().Name(), "Field:",
						reflect.TypeOf(v).Elem().Field(i).Name, "has tag `inject` but value is not `true`")
				}
				for _, candi := range components {
					// have tag with inject
					if reflect.ValueOf(candi).Type().AssignableTo(cElement.Field(i).Type()) {
						cElement.Field(i).Set(reflect.ValueOf(candi))
						logger.Info("Component:", cElement.Type().Name(), "Field:",
							reflect.TypeOf(v).Elem().Field(i).Name, "has been injected with component",
							reflect.ValueOf(candi).Elem().Type().Name())
						break
					}

				}

			}

		}

	}
}
