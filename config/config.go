/*
 * Copyright 2019 Azz. All rights reserved.
 * Use of this source code is governed by a GPL-3.0
 * license that can be found in the LICENSE file.
 */

package config

import (
	"github.com/azzill/goze/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

var logger = log.NewLogger("Configuration")

type CommonConfiguration struct {
	configs map[interface{}]interface{}
}

func (c *CommonConfiguration) DefaultGet(path string, def interface{}) interface{} {
	get := c.Get(path)
	if get == nil {
		return def
	} else {
		return get
	}
}

func (c *CommonConfiguration) Get(path string) interface{} {
	keys := strings.Split(path, ".")
	r := c.configs
	for i := range keys {
		if r[keys[i]] == nil {
			return nil
		}
		if i == len(keys)-1 {
			return r[keys[i]]
		}
		r = r[keys[i]].(map[interface{}]interface{})
		if r == nil {
			return nil
		}
	}
	return nil
}

func NewCommonConfiguration(y string) *CommonConfiguration {
	config := &CommonConfiguration{configs: map[interface{}]interface{}{}}
	b, err := ioutil.ReadFile(y)
	if err != nil {
		logger.Error("Error reading config file", err)
	}
	err = yaml.Unmarshal(b, &config.configs)
	if err != nil {
		logger.Error("Error parsing config file", err)
	}
	return config
}
